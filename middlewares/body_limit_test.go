package middlewares

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pionus/arry"
)

func TestBodyLimitRejectsOversizedBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(strings.Repeat("a", 100)))
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	var readErr error
	BodyLimit(10)(func(ctx arry.Context) {
		_, readErr = io.ReadAll(ctx.Request().Body)
	})(ctx)

	if readErr == nil {
		t.Fatal("expected read error for oversized body")
	}
}

func TestBodyLimitAllowsNormalBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"ok"}`))
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	var got string
	BodyLimit(1024)(func(ctx arry.Context) {
		var payload struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(ctx.Request().Body).Decode(&payload); err != nil {
			t.Errorf("decode: %v", err)
			return
		}
		got = payload.Name
	})(ctx)

	if got != "ok" {
		t.Fatalf("name = %q", got)
	}
}

func TestAuthFunc(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	AuthFunc(func(ctx arry.Context) bool {
		ctx.Set("user", "alice")
		return true
	})(func(ctx arry.Context) {
		called = true
		if got := ctx.Get("user"); got != "alice" {
			t.Fatalf("user = %v", got)
		}
	})(ctx)

	if !called {
		t.Fatal("handler should run")
	}
}

func TestAuthFuncGuardsMethod(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	AuthFunc(func(ctx arry.Context) bool { return false }, http.MethodPost)(
		func(ctx arry.Context) { called = true },
	)(ctx)

	if called {
		t.Fatal("handler should not run")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}
