package middlewares

import (
	"net/http"

	"github.com/pionus/arry"
)

// BodyLimit caps the request body size. Requests exceeding the limit fail when
// the handler reads the body (for example during Decode or ParseForm).
func BodyLimit(maxBytes int64) arry.Middleware {
	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			if ctx.Request().Body != nil {
				ctx.Request().Body = http.MaxBytesReader(ctx.Response(), ctx.Request().Body, maxBytes)
			}
			next(ctx)
		}
	}
}
