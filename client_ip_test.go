package arry

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func clientIPRequest(a *Arry, remote, xff, realIP string) string {
	req := httptest.NewRequest(http.MethodGet, "/ip", nil)
	req.RemoteAddr = remote
	if xff != "" {
		req.Header.Set("X-Forwarded-For", xff)
	}
	if realIP != "" {
		req.Header.Set("X-Real-IP", realIP)
	}
	rec := httptest.NewRecorder()
	a.Router().Get("/ip", func(ctx Context) { ctx.Text(http.StatusOK, ctx.ClientIP()) })
	a.ServeHTTP(rec, req)
	return rec.Body.String()
}

func TestClientIPIgnoresXFFWithoutTrustedProxy(t *testing.T) {
	a := New()
	if got := clientIPRequest(a, "1.2.3.4:80", "9.9.9.9", ""); got != "1.2.3.4" {
		t.Fatalf("ip = %q", got)
	}
}

func TestClientIPTrustsProxyAndUsesRightmostUntrusted(t *testing.T) {
	a := New()
	if err := a.SetTrustedProxies([]string{"10.0.0.0/8"}); err != nil {
		t.Fatal(err)
	}
	if got := clientIPRequest(a, "10.1.2.3:80", "1.1.1.1, 8.8.8.8", ""); got != "8.8.8.8" {
		t.Fatalf("ip = %q", got)
	}
	if got := clientIPRequest(a, "10.1.2.3:80", "8.8.8.8, 10.0.0.2", ""); got != "8.8.8.8" {
		t.Fatalf("chained ip = %q", got)
	}
	if got := clientIPRequest(a, "10.1.2.3:80", "", "7.7.7.7"); got != "7.7.7.7" {
		t.Fatalf("real ip = %q", got)
	}
}

func TestClientIPIgnoresXFFFromUntrustedPeer(t *testing.T) {
	a := New()
	if err := a.SetTrustedProxies([]string{"10.0.0.0/8"}); err != nil {
		t.Fatal(err)
	}
	if got := clientIPRequest(a, "1.2.3.4:80", "9.9.9.9", ""); got != "1.2.3.4" {
		t.Fatalf("ip = %q", got)
	}
}

func TestSetTrustedProxiesRejectsInvalid(t *testing.T) {
	a := New()
	if err := a.SetTrustedProxies([]string{"not-an-ip"}); err == nil {
		t.Fatal("expected invalid proxy error")
	}
}
