package middlewares

import (
    "io"
    "net/http"
    "compress/gzip"
    "github.com/pionus/arry"
)


type GzipWriter struct {
    io.Writer
    http.ResponseWriter
}

// func (w *GzipWriter) Header() http.Header {
//     return w.ResponseWriter.Header()
// }
//
func (w *GzipWriter) Write(b []byte) (int, error) {
    if w.Header().Get("Content-Type") == "" {
        w.Header().Set("Content-Type", http.DetectContentType(b))
    }
    return w.Writer.Write(b)
}


func Gzip(next arry.Handler) arry.Handler {
    return func(ctx arry.Context) {
        rw := ctx.Response().Writer
        w, err := gzip.NewWriterLevel(rw, 5)
        if err != nil {
            panic("cannot get gzip writer!")
        }

        defer w.Close()

        gw := GzipWriter{
            Writer: w,
            ResponseWriter: rw,
        }
        
        gw.Header().Set("Content-Encoding", "gzip")
        ctx.Response().Writer = &gw

        next(ctx)
    }
}
