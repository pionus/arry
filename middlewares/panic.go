package middlewares

import (
	"log/slog"

	"github.com/pionus/arry"
)

// PanicHandler is a function that handles a recovered panic value.
type PanicHandler func(ctx arry.Context, err interface{})

// Panic is the default panic recovery middleware.
// It recovers from panics and returns a 500 status using ctx.Reply().
func Panic(next arry.Handler) arry.Handler {
	return func(ctx arry.Context) {
		defer panicHandler(ctx)
		next(ctx)
	}
}

// PanicWithHandler returns a panic recovery middleware that uses
// the given handler function to produce the error response.
func PanicWithHandler(h PanicHandler) arry.Middleware {
	return func(next arry.Handler) arry.Handler {
		return func(ctx arry.Context) {
			defer func() {
				if e := recover(); e != nil {
					slog.Error("panic recovered", "error", e)
					h(ctx, e)
				}
			}()
			next(ctx)
		}
	}
}

func panicHandler(ctx arry.Context) {
	if e := recover(); e != nil {
		ctx.Reply(500)
	}
}
