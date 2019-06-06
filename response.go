package arry

import (
    "net/http"
)

type Response struct {
    Writer http.ResponseWriter
    Code int
    Sent bool
}


func (r *Response) Header() http.Header {
    return r.Writer.Header()
}

func (r *Response) Write(b []byte) (int, error) {
    if !r.Sent {
        r.WriteHeader(r.Code)
    }

    return r.Writer.Write(b)
}

func (r *Response) WriteHeader(code int) {
    r.Writer.WriteHeader(code)
    r.Sent = true
}
