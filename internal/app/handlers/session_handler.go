package handlers

import (
	"context"
	"github.com/iota-agency/iota-erp/internal/app/services"
	"github.com/iota-agency/iota-erp/internal/domain/entities/authlog"
	"github.com/iota-agency/iota-erp/internal/domain/entities/session"
	"github.com/iota-agency/iota-erp/sdk/composables"
	"github.com/iota-agency/iota-erp/sdk/event"
	"gorm.io/gorm"
	"log"
	"time"
)

type SessionEventsHandler struct {
	db             *gorm.DB
	publisher      event.Publisher
	authLogService *services.AuthLogService
}

func RegisterSessionEventHandlers(
	db *gorm.DB,
	publisher event.Publisher,
	authLogService *services.AuthLogService,
) *SessionEventsHandler {
	handler := &SessionEventsHandler{
		db:             db,
		publisher:      publisher,
		authLogService: authLogService,
	}
	publisher.Subscribe(handler.onSessionCreated)
	return handler
}

func (h *SessionEventsHandler) onSessionCreated(event session.CreatedEvent) {
	sess := event.Result
	logEntity := &authlog.AuthenticationLog{
		ID:        0,
		UserID:    sess.UserID,
		IP:        sess.IP,
		UserAgent: sess.UserAgent,
		CreatedAt: time.Now(),
	}
	tx := h.db.Begin()
	ctx := composables.WithTx(context.Background(), tx)
	if err := h.authLogService.Create(ctx, logEntity); err != nil {
		log.Fatalf("failed to create auth log: %v", err)
	}
	if err := tx.Commit().Error; err != nil {
		log.Fatalf("failed to commit transaction: %v", err)
	}
}
