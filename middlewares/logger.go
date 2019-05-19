package middlewares

import (
    "time"
    "github.com/pionus/arry"
    "log"
)

func Logger(next arry.Handler) arry.Handler {
    return func(ctx arry.Context) {
        start := time.Now()

        req := ctx.Request()

        path := req.URL.Path
        ua := req.Header.Get("User-Agent")

        log.Printf("%s|%s|%s|%s\n", req.RemoteAddr, req.Method, path, ua)

        next(ctx)

        delta := time.Now().Sub(start)
        log.Printf("%d|%s|%dns\n", ctx.GetStatus(), path, delta.Nanoseconds())
    }
}
