package constants

import "github.com/go-playground/validator/v10"

type ContextKey string

const (
	UserKey     ContextKey = "user"
	SessionKey  ContextKey = "session"
	NavItemsKey ContextKey = "navItems"
	TxKey       ContextKey = "tx"
	ParamsKey   ContextKey = "params"
	LoggerKey   ContextKey = "logger"
)

var Validate = validator.New(validator.WithRequiredStructEnabled())
