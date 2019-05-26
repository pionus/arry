package middlewares

import (
    "github.com/pionus/arry"
)

func Panic(next arry.Handler) arry.Handler {
    return func(ctx arry.Context) {
        defer panicHandler(ctx)
        next(ctx)
    }
}

func panicHandler(ctx arry.Context) {
	if e := recover(); e != nil {
        ctx.Reply(500)
	}
}
