package middleware

import (
	"github.com/iota-uz/iota-sdk/pkg/composables"
	"github.com/iota-uz/iota-sdk/pkg/constants"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

func WithLocalizer(bundle *i18n.Bundle) mux.MiddlewareFunc {
	return ContextKeyValue(
		constants.LocalizerKey,
		func(r *http.Request, _ http.ResponseWriter) interface{} {
			locale := composables.UseLocale(r.Context(), language.English)
			return i18n.NewLocalizer(bundle, locale.String())
		},
	)
}
