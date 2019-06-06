package middlewares

import (
    "testing"
    "net/http/httptest"

    "github.com/pionus/arry"
)


func TestPanic(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()

    ctx := arry.NewContext(req, rec)

    h := Panic(func(ctx arry.Context) {
        panic(200)
    })

    h(ctx)

    if rec.Code != 500 {
        t.Errorf("panic is failed, %d", rec.Code)
    }


}
