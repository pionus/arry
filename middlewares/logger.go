package middlewares

import (
    "io"
    "log"
    "os"
    "path/filepath"
    "time"

    "github.com/pionus/arry"
)

func Logger() arry.Middleware {
    return LoggerToWriter(os.Stdout)
}


func LoggerToFile(file string) arry.Middleware {
    if dir := filepath.Dir(file); dir != "." {
        os.MkdirAll(dir, 0755)
    }

    f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Fatal(err)
    }

    return LoggerToWriter(f)
}


func LoggerToWriter(out io.Writer) arry.Middleware {
    logger := log.New(out, "", log.LstdFlags)
    file, isFile := out.(*os.File)

    return func(next arry.Handler) arry.Handler {
        return func(ctx arry.Context) {
            start := time.Now()

            req := ctx.Request()

            path := req.URL.Path
            ua := req.Header.Get("User-Agent")

            logger.Printf("%s|%s|%s|%s\n", req.RemoteAddr, req.Method, path, ua)

            next(ctx)

            delta := time.Now().Sub(start)
            logger.Printf("%d|%s|%dμs\n", ctx.Response().Code, path, delta.Microseconds())

            if isFile {
                file.Sync()
            }
        }
    }
}
