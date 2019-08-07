package middlewares

import (
    "testing"
    "net/http/httptest"

    "github.com/pionus/arry"
)


func TestAuthSussess(t *testing.T) {
    token := "token123"

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", token)

    rec := httptest.NewRecorder()

    ctx := arry.NewContext(req, rec)

    auth := false
    h := Auth(token)(func(ctx arry.Context) {
        auth = ctx.Get("auth").(bool)
        ctx.Text(200, "OK")
    })

    h(ctx)

    if rec.Code != 200 || !auth {
        t.Errorf("logger failed, %s", rec.Body.String())
    }
}


func TestAuthFailed(t *testing.T) {
    token := "token123"

    req := httptest.NewRequest("GET", "/", nil)
    req.Header.Set("Authorization", "something")

    rec := httptest.NewRecorder()

    ctx := arry.NewContext(req, rec)

    auth := false
    h := Auth(token)(func(ctx arry.Context) {
        auth = ctx.Get("auth").(bool)
        ctx.Text(200, "OK")
    })

    h(ctx)

    if rec.Code != 200 || auth {
        t.Errorf("logger failed, %s", rec.Body.String())
    }
}
