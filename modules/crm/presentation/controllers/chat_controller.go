package controllers

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"sync"

	"github.com/a-h/templ"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/iota-uz/iota-sdk/modules/core/domain/entities/phone"
	"github.com/iota-uz/iota-sdk/modules/crm/domain/aggregates/chat"
	"github.com/iota-uz/iota-sdk/modules/crm/domain/aggregates/client"
	"github.com/iota-uz/iota-sdk/modules/crm/infrastructure/persistence"
	"github.com/iota-uz/iota-sdk/modules/crm/presentation/mappers"
	"github.com/iota-uz/iota-sdk/modules/crm/presentation/templates/pages/chats"
	"github.com/iota-uz/iota-sdk/modules/crm/presentation/viewmodels"
	"github.com/iota-uz/iota-sdk/modules/crm/services"
	"github.com/iota-uz/iota-sdk/pkg/application"
	"github.com/iota-uz/iota-sdk/pkg/composables"
	"github.com/iota-uz/iota-sdk/pkg/mapping"
	"github.com/iota-uz/iota-sdk/pkg/middleware"
	"github.com/iota-uz/iota-sdk/pkg/shared"
)

type CreateChatDTO struct {
	Phone string
}

type SendMessageDTO struct {
	Message string
}

type ChatController struct {
	app             application.Application
	wsHandler       *WebSocketHandler
	templateService *services.MessageTemplateService
	clientService   *services.ClientService
	chatService     *services.ChatService
	basePath        string
}

func NewChatController(app application.Application, basePath string) application.Controller {
	return &ChatController{
		app:             app,
		wsHandler:       NewWebSocketHandler(),
		clientService:   app.Service(services.ClientService{}).(*services.ClientService),
		chatService:     app.Service(services.ChatService{}).(*services.ChatService),
		templateService: app.Service(services.MessageTemplateService{}).(*services.MessageTemplateService),
		basePath:        basePath,
	}
}

func (c *ChatController) Key() string {
	return c.basePath
}

func (c *ChatController) Register(r *mux.Router) {
	commonMiddleware := []mux.MiddlewareFunc{
		middleware.Authorize(),
		middleware.RedirectNotAuthenticated(),
		middleware.ProvideUser(),
		middleware.Tabs(),
		middleware.WithLocalizer(c.app.Bundle()),
		middleware.NavItems(),
		middleware.WithPageContext(),
	}
	getRouter := r.PathPrefix(c.basePath).Subrouter()
	getRouter.Use(commonMiddleware...)
	getRouter.HandleFunc("", c.List).Methods(http.MethodGet)
	getRouter.HandleFunc("/new", c.GetNew).Methods(http.MethodGet)
	getRouter.HandleFunc("/ws", c.wsHandler.ServeHTTP)

	setRouter := r.PathPrefix(c.basePath).Subrouter()
	setRouter.Use(commonMiddleware...)
	setRouter.Use(middleware.WithTransaction())
	setRouter.HandleFunc("", c.Create).Methods(http.MethodPost)
	setRouter.HandleFunc("/{id:[0-9]+}/messages", c.SendMessage).Methods(http.MethodPost)
}

func (c *ChatController) messageTemplates(ctx context.Context) ([]*viewmodels.MessageTemplate, error) {
	templates, err := c.templateService.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	return mapping.MapViewModels(templates, mappers.MessageTemplateToViewModel), nil
}

func (c *ChatController) List(w http.ResponseWriter, r *http.Request) {
	chatEntities, err := c.chatService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chatViewModels := mapping.MapViewModels(chatEntities, mappers.ChatToViewModel)
	messageTemplates, err := c.messageTemplates(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	props := &chatsui.IndexPageProps{
		WebsocketURL: c.basePath + "/ws",
		ClientsURL:   "/crm/clients",
		NewChatURL:   "/crm/chats/new",
		Chats:        chatViewModels,
	}
	templHandler := templ.Handler(
		chatsui.Index(props),
		templ.WithStreaming(),
	)
	ctx := r.Context()
	chatID := r.URL.Query().Get("chat_id")
	if chatID == "" {
		templHandler.ServeHTTP(
			w, r.WithContext(templ.WithChildren(ctx, chatsui.NoSelectedChat())),
		)
		return
	}
	for _, chat := range chatViewModels {
		if chat.ID == chatID {
			props := chatsui.SelectedChatProps{
				BaseURL:    c.basePath,
				ClientsURL: "/crm/clients",
				Chat:       chat,
				Templates:  messageTemplates,
			}
			templHandler.ServeHTTP(
				w, r.WithContext(templ.WithChildren(ctx, chatsui.SelectedChat(props))),
			)
			return
		}
	}
	templHandler.ServeHTTP(
		w, r.WithContext(templ.WithChildren(ctx, chatsui.ChatNotFound())),
	)
}

func (c *ChatController) GetNew(w http.ResponseWriter, r *http.Request) {
	chatEntities, err := c.chatService.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	chatViewModels := mapping.MapViewModels(chatEntities, mappers.ChatToViewModel)
	ctx := r.Context()
	props := &chatsui.IndexPageProps{
		Chats: chatViewModels,
	}
	templHandler := templ.Handler(chatsui.Index(props), templ.WithStreaming())
	templHandler.ServeHTTP(
		w, r.WithContext(templ.WithChildren(ctx, chatsui.NewChat(chatsui.NewChatProps{
			BaseURL:       c.basePath,
			CreateChatURL: c.basePath,
			Phone:         "+1",
			Errors:        map[string]string{},
		}))),
	)
}

func (c *ChatController) Create(w http.ResponseWriter, r *http.Request) {
	dto, err := composables.UseForm(&CreateChatDTO{}, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	createdClient, err := c.clientService.Create(r.Context(), &client.CreateDTO{
		FirstName: "Unknown",
		LastName:  "Unknown",
		Phone:     dto.Phone,
	})
	if errors.Is(err, phone.ErrInvalidPhoneNumber) {
		templ.Handler(chatsui.NewChatForm(chatsui.NewChatProps{
			BaseURL:       c.basePath,
			CreateChatURL: c.basePath,
			Phone:         dto.Phone,
			Errors: map[string]string{
				"Phone": err.Error(),
			},
		})).ServeHTTP(w, r)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, err = c.chatService.Create(r.Context(), &chat.CreateDTO{
		ClientID: createdClient.ID(),
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	shared.HxRedirect(w, r, c.basePath)
}

func (c *ChatController) SendMessage(w http.ResponseWriter, r *http.Request) {
	chatID, err := shared.ParseID(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	dto, err := composables.UseForm(&SendMessageDTO{}, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	updatedChat, err := c.chatService.SendMessage(r.Context(), chatID, dto.Message)
	if errors.Is(err, persistence.ErrChatNotFound) {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	messageTemplates, err := c.messageTemplates(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	messageViewModels := mapping.MapViewModels(updatedChat.Messages(), mappers.MessageToViewModel)
	chatsui.ChatMessages(
		messageViewModels,
		mappers.ClientToViewModel(updatedChat.Client()),
	).Render(r.Context(), &buf)
	c.wsHandler.Broadcast(websocket.TextMessage, buf.Bytes())
	props := chatsui.SelectedChatProps{
		BaseURL:    c.basePath,
		ClientsURL: "/crm/clients",
		Chat:       mappers.ChatToViewModel(updatedChat),
		Templates:  messageTemplates,
	}
	component := chatsui.SelectedChat(props)
	templ.Handler(component).ServeHTTP(w, r)
}

// WebSocketHandler handles WebSocket connections
type WebSocketHandler struct {
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.RWMutex
}

// NewWebSocketHandler creates a new WebSocket handler
func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		clients: make(map[*websocket.Conn]bool),
	}
}

// readPump reads messages from the WebSocket connection
func (h *WebSocketHandler) readPump(conn *websocket.Conn) {
	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		conn.Close()
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("Error reading message: %v", err)
			}
			return
		}

		if err := h.handleMessage(conn, messageType, message); err != nil {
			log.Printf("Error handling message: %v", err)
			return
		}
	}
}

// ServeHTTP implements the http.Handler interface
func (h *WebSocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	h.mu.Lock()
	h.clients[conn] = true
	h.mu.Unlock()

	// Start reading messages in a separate goroutine and return
	go h.readPump(conn)
}

// handleMessage processes incoming messages
func (h *WebSocketHandler) handleMessage(_ *websocket.Conn, _ int, _ []byte) error {
	return nil
}

// broadcast sends a message to all connected clients except the sender
func (h *WebSocketHandler) Broadcast(messageType int, message []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for client := range h.clients {
		if err := client.WriteMessage(messageType, message); err != nil {
			log.Printf("Error broadcasting message: %v", err)
		}
	}
}
