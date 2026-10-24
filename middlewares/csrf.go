package middlewares

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/pionus/arry"
)

// CSRFContextKey is the context key holding the current CSRF token.
const CSRFContextKey = "csrf_token"

// CSRFConfig configures the CSRF middleware.
type CSRFConfig struct {
	// CookieName is the name of the CSRF cookie. Defaults to "csrf_token".
	CookieName string
	// FormField is the form field carrying the token. Defaults to "_csrf".
	FormField string
	// HeaderName is the header carrying the token. Defaults to "X-CSRF-Token".
	HeaderName string
	// Path is the cookie path. Defaults to "/".
	Path string
	// Secure marks the cookie Secure. Enable it in production behind HTTPS.
	Secure bool
	// SameSite defaults to http.SameSiteLaxMode.
	SameSite http.SameSite
	// TokenBytes is the number of random bytes. Defaults to 32.
	TokenBytes int
	// MaxAge is the cookie lifetime in seconds. Zero means a session cookie.
	MaxAge int
}

// CSRF protects unsafe requests using the double-submit cookie pattern.
// Safe methods ensure a token exists and expose it through the context key
// CSRFContextKey; unsafe methods compare the form field or header against the
// cookie.
func CSRF() arry.Middleware {
	return CSRFWithConfig(CSRFConfig{})
}

// CSRFWithConfig returns a CSRF middleware with the given configuration.
func CSRFWithConfig(config CSRFConfig) arry.Middleware {
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.FormField == "" {
		config.FormField = "_csrf"
	}
	if config.HeaderName == "" {
		config.HeaderName = "X-CSRF-Token"
	}
	if config.Path == "" {
		config.Path = "/"
	}
	if config.SameSite == 0 {
		config.SameSite = http.SameSiteLaxMode
	}
	if config.TokenBytes <= 0 {
		config.TokenBytes = 32
	}

	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			token := ""
			if cookie, err := ctx.Request().Cookie(config.CookieName); err == nil {
				token = cookie.Value
			}
			if token == "" {
				token = newToken(config.TokenBytes)
				ctx.SetCookie(&http.Cookie{
					Name:     config.CookieName,
					Value:    token,
					Path:     config.Path,
					MaxAge:   config.MaxAge,
					Secure:   config.Secure,
					HttpOnly: true,
					SameSite: config.SameSite,
				})
			}
			ctx.Set(CSRFContextKey, token)

			if isSafeMethod(ctx.Request().Method) {
				next(ctx)
				return
			}

			got := ctx.Header(config.HeaderName)
			if got == "" {
				got = ctx.Request().PostFormValue(config.FormField)
			}

			if subtle.ConstantTimeCompare([]byte(got), []byte(token)) != 1 {
				ctx.Reply(http.StatusForbidden)
				return
			}

			next(ctx)
		}
	}
}

// CSRFToken returns the token attached by the CSRF middleware.
func CSRFToken(ctx arry.Context) string {
	token, _ := ctx.Get(CSRFContextKey).(string)
	return token
}

func isSafeMethod(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodGet, http.MethodHead, http.MethodOptions:
		return true
	}
	return false
}

func newToken(n int) string {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		panic("arry: cannot generate CSRF token: " + err.Error())
	}
	return hex.EncodeToString(buf)
}
