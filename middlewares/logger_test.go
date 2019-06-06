package middlewares

import (
    "testing"
    "net/http/httptest"

    "github.com/pionus/arry"
)


func TestLogger(t *testing.T) {
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()

    ctx := arry.NewContext(req, rec)

    h := Logger(func(ctx arry.Context) {
        ctx.Text(200, "OK")
    })

    h(ctx)

    if rec.Code != 200 {
        t.Errorf("logger failed, %s", rec.Body.String())
    }


}
