package middlewares

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/pionus/arry"
)

func TestGzipCompressesLargeText(t *testing.T) {
	body := strings.Repeat("hello world ", 200)

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	Gzip(func(ctx arry.Context) {
		ctx.Text(200, body)
	})(ctx)

	if enc := rec.Header().Get("Content-Encoding"); enc != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", enc)
	}

	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	got, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != body {
		t.Fatalf("decoded body mismatch: got %d bytes", len(got))
	}
}

func TestGzipSkipsSmallResponse(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	Gzip(func(ctx arry.Context) {
		ctx.Text(200, "OK")
	})(ctx)

	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Content-Encoding = %q, want empty", enc)
	}
	if rec.Body.String() != "OK" {
		t.Fatalf("body = %q", rec.Body.String())
	}
}

func TestGzipSkipsWhenNotAccepted(t *testing.T) {
	body := bytes.Repeat([]byte("a"), 4096)

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	Gzip(func(ctx arry.Context) {
		ctx.Text(200, string(body))
	})(ctx)

	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Content-Encoding = %q, want empty", enc)
	}
}

func TestGzipSkipsHEAD(t *testing.T) {
	req := httptest.NewRequest("HEAD", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	Gzip(func(ctx arry.Context) {
		ctx.Text(200, strings.Repeat("a", 4096))
	})(ctx)

	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Content-Encoding = %q, want empty", enc)
	}
}

func TestGzipNoContent(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Accept-Encoding", "gzip")
	rec := httptest.NewRecorder()
	ctx := arry.NewContext(req, rec)

	Gzip(func(ctx arry.Context) {
		ctx.NoContent(204)
	})(ctx)

	if rec.Code != 204 {
		t.Fatalf("status = %d, want 204", rec.Code)
	}
	if enc := rec.Header().Get("Content-Encoding"); enc != "" {
		t.Fatalf("Content-Encoding = %q, want empty", enc)
	}
}
