package middlewares

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/pionus/arry"
)

// Auth returns an authentication middleware that compares the request token
// with the expected token using a constant-time comparison.
//
// The token is read from the Authorization header. Both a bare token
// ("Authorization: <token>") and the Bearer scheme ("Authorization: Bearer
// <token>") are accepted.
//
// The optional methods list guards only those HTTP methods. If the request
// method is in the list and the request is unauthenticated, the middleware
// responds 403 and short-circuits. Requests with other methods pass through
// with the auth state set, so handlers can still inspect it. With no methods,
// every request passes through.
func Auth(token string, methods ...string) arry.Middleware {
	guardSet := make(map[string]bool, len(methods))
	for _, m := range methods {
		guardSet[strings.ToUpper(m)] = true
	}

	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			got := requestToken(ctx.Request())
			isAuthed := token != "" && subtle.ConstantTimeCompare([]byte(got), []byte(token)) == 1
			ctx.Set("auth", isAuthed)

			if len(guardSet) > 0 && guardSet[ctx.Request().Method] && !isAuthed {
				ctx.JSONError(http.StatusForbidden, "forbidden")
				return
			}

			next(ctx)
		}
	}
}

func requestToken(r *http.Request) string {
	raw := strings.TrimSpace(r.Header.Get("Authorization"))
	if len(raw) > 7 && strings.EqualFold(raw[:7], "bearer ") {
		return strings.TrimSpace(raw[7:])
	}
	return raw
}
