package middlewares

import (
	"net/http"

	"github.com/pionus/arry"
)

// Auth returns an authentication middleware.
//   - token: expected Authorization header value
//   - methods: HTTP methods to guard (e.g. []string{"POST", "PUT", "DELETE"}).
//     If the request method is in this list and unauthenticated, responds 403 and short-circuits.
//     If the request method is NOT in this list, sets auth state but allows the request through.
//     If methods is nil or empty, all requests are allowed through (backward compatible).
func Auth(token string, methods []string) arry.Middleware {
	guardSet := make(map[string]bool, len(methods))
	for _, m := range methods {
		guardSet[m] = true
	}

	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			auth := ctx.Request().Header.Get("Authorization")
			isAuthed := auth != "" && auth == token
			ctx.Set("auth", isAuthed)

			if len(guardSet) > 0 && guardSet[ctx.Request().Method] && !isAuthed {
				ctx.JSONError(http.StatusForbidden, "forbidden")
				return
			}

			next(ctx)
		}
	}
}
