package middlewares

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pionus/arry"
)

// CORSConfig defines the configuration for the CORS middleware.
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           time.Duration
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
	exposeHeaders := strings.Join(config.ExposeHeaders, ", ")

	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			origin := ctx.Header("Origin")
			if origin == "" {
				next(ctx)
				return
			}

			wildcard := false
			allowed := false
			for _, o := range config.AllowOrigins {
				if o == "*" {
					wildcard = true
					allowed = true
					break
				}
				if o == origin {
					allowed = true
					break
				}
			}

			if !allowed {
				next(ctx)
				return
			}

			if wildcard && !config.AllowCredentials {
				ctx.SetHeader("Access-Control-Allow-Origin", "*")
			} else {
				ctx.SetHeader("Access-Control-Allow-Origin", origin)
				addVary(ctx.Response().Header(), "Origin")
			}

			if config.AllowCredentials {
				ctx.SetHeader("Access-Control-Allow-Credentials", "true")
			}
			if exposeHeaders != "" {
				ctx.SetHeader("Access-Control-Expose-Headers", exposeHeaders)
			}

			if ctx.Request().Method == http.MethodOptions && ctx.Header("Access-Control-Request-Method") != "" {
				ctx.SetHeader("Access-Control-Allow-Methods", allowMethods)
				ctx.SetHeader("Access-Control-Allow-Headers", allowHeaders)
				addVary(ctx.Response().Header(), "Access-Control-Request-Method")
				addVary(ctx.Response().Header(), "Access-Control-Request-Headers")
				if config.MaxAge > 0 {
					ctx.SetHeader("Access-Control-Max-Age", strconv.Itoa(int(config.MaxAge/time.Second)))
				}
				ctx.NoContent(http.StatusNoContent)
				return
			}

			next(ctx)
		}
	}
}
