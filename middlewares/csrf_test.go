package middlewares

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pionus/arry"
)

func TestCSRFSafeRequestSetsToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	var token string
	CSRF()(func(ctx arry.Context) {
		token = CSRFToken(ctx)
		ctx.Text(http.StatusOK, "ok")
	})(ctx)

	if token == "" {
		t.Fatal("missing CSRF token in context")
	}
	cookie := findCookie(rec, "csrf_token")
	if cookie == nil || cookie.Value != token {
		t.Fatalf("cookie = %+v", cookie)
	}
	if !cookie.HttpOnly {
		t.Fatal("CSRF cookie should be HttpOnly")
	}
}

func TestCSRFRejectsMissingToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	CSRF()(func(ctx arry.Context) { called = true })(ctx)

	if called {
		t.Fatal("handler should not run")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func TestCSRFAcceptsHeader(t *testing.T) {
	token := "abc123"
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	req.Header.Set("X-CSRF-Token", token)
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	CSRF()(func(ctx arry.Context) {
		called = true
		ctx.Text(http.StatusOK, "ok")
	})(ctx)

	if !called {
		t.Fatal("handler should run with a valid header token")
	}
}

func TestCSRFAcceptsFormField(t *testing.T) {
	token := "abc123"
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("_csrf="+token))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	CSRF()(func(ctx arry.Context) {
		called = true
		ctx.Text(http.StatusOK, "ok")
	})(ctx)

	if !called {
		t.Fatal("handler should run with a valid form token")
	}
}

func TestCSRFIgnoresQueryToken(t *testing.T) {
	token := "abc123"
	req := httptest.NewRequest(http.MethodPost, "/?_csrf="+token, nil)
	req.AddCookie(&http.Cookie{Name: "csrf_token", Value: token})
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	called := false
	CSRF()(func(ctx arry.Context) { called = true })(ctx)

	if called {
		t.Fatal("query token must not satisfy CSRF")
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", rec.Code)
	}
}

func findCookie(rec *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}
