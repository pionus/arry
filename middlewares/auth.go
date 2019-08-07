package middlewares

import (
    "github.com/pionus/arry"
)


func Auth(token string) arry.Middleware {
    return func(next arry.Handler) arry.Handler {
        return func(ctx arry.Context) {
            auth := ctx.Request().Header.Get("Authorization")
            ctx.Set("auth", auth != "" && auth == token)

            next(ctx)
        }
    }
}
