package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pionus/arry"
)

func TestCORSPreflight(t *testing.T) {
	req := httptest.NewRequest(http.MethodOptions, "/api/posts", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	CORS()(func(ctx arry.Context) {
		t.Fatal("preflight should not reach the handler")
	})(ctx)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Allow-Origin = %q, want *", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Fatal("missing Access-Control-Allow-Methods")
	}
}

func TestCORSCredentialsEchoesOrigin(t *testing.T) {
	cfg := CORSConfig{
		AllowOrigins:     []string{"https://example.com"},
		AllowCredentials: true,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://example.com")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	CORSWithConfig(cfg)(func(ctx arry.Context) {
		ctx.Text(http.StatusOK, "OK")
	})(ctx)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Fatalf("Allow-Origin = %q", got)
	}
	if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Fatalf("Allow-Credentials = %q", got)
	}
	if got := rec.Header().Get("Vary"); got != "Origin" {
		t.Fatalf("Vary = %q", got)
	}
}

func TestCORSDisallowedOrigin(t *testing.T) {
	cfg := CORSConfig{AllowOrigins: []string{"https://example.com"}}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "https://evil.test")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	CORSWithConfig(cfg)(func(ctx arry.Context) {
		ctx.Text(http.StatusOK, "OK")
	})(ctx)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("Allow-Origin = %q, want empty", got)
	}
}
