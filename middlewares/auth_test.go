package middlewares

import (
	"net/http/httptest"
	"testing"

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

func TestAuthBearer(t *testing.T) {
	token := "token123"

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	ctx := arry.NewContext(req, rec)
	auth := false
	Auth(token)(func(ctx arry.Context) {
		auth = ctx.IsAuthed()
		ctx.Text(200, "OK")
	})(ctx)

	if !auth {
		t.Fatal("Bearer token should authenticate")
	}
}

func TestAuthGuardedMethod(t *testing.T) {
	token := "token123"

	req := httptest.NewRequest("POST", "/", nil)
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	Auth(token, "POST")(func(ctx arry.Context) {
		called = true
	})(ctx)

	if called {
		t.Fatal("handler should not run for unauthenticated guarded method")
	}
	if rec.Code != 403 {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestAuthUnguardedMethodSetsState(t *testing.T) {
	token := "token123"

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	Auth(token, "POST")(func(ctx arry.Context) {
		called = true
		if ctx.IsAuthed() {
			t.Fatal("request should not be authenticated")
		}
	})(ctx)

	if !called {
		t.Fatal("handler should run for unguarded method")
	}
}
