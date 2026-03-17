package middlewares

import (
	"net/http"
	"strings"

	"github.com/pionus/arry"
)

// CORSConfig defines the configuration for the CORS middleware.
type CORSConfig struct {
	AllowOrigins []string
	AllowMethods []string
	AllowHeaders []string
}

// DefaultCORSConfig returns a permissive default CORS configuration.
var DefaultCORSConfig = CORSConfig{
	AllowOrigins: []string{"*"},
	AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions},
	AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
}

// CORS returns a CORS middleware with default configuration.
func CORS() arry.Middleware {
	return CORSWithConfig(DefaultCORSConfig)
}

// CORSWithConfig returns a CORS middleware with the given configuration.
func CORSWithConfig(config CORSConfig) arry.Middleware {
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = DefaultCORSConfig.AllowOrigins
	}
	if len(config.AllowMethods) == 0 {
		config.AllowMethods = DefaultCORSConfig.AllowMethods
	}
	if len(config.AllowHeaders) == 0 {
		config.AllowHeaders = DefaultCORSConfig.AllowHeaders
	}

	allowMethods := strings.Join(config.AllowMethods, ", ")
	allowHeaders := strings.Join(config.AllowHeaders, ", ")

	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			origin := ctx.Header("Origin")
			if origin == "" {
				next(ctx)
				return
			}

			allowed := false
			for _, o := range config.AllowOrigins {
				if o == "*" || o == origin {
					allowed = true
					break
				}
			}

			if !allowed {
				next(ctx)
				return
			}

			ctx.SetHeader("Access-Control-Allow-Origin", origin)
			if config.AllowOrigins[0] == "*" {
				ctx.SetHeader("Access-Control-Allow-Origin", "*")
			}
			ctx.SetHeader("Access-Control-Allow-Methods", allowMethods)
			ctx.SetHeader("Access-Control-Allow-Headers", allowHeaders)

			// Handle preflight
			if ctx.Request().Method == http.MethodOptions {
				ctx.Status(http.StatusNoContent)
				return
			}

			next(ctx)
		}
	}
}
