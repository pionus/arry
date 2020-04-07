package arry

import (
	"testing"
    // "net/http/httptest"
)


func TestRouter(t *testing.T) {
	router := NewRouter()
	
	node := router.Route("/", &context{})

	if node == nil {
		t.Error("default router is not correct")
	}
}

func TestRouterPath(t *testing.T) {
	router := NewRouter()
	url := "/path/to/route"

	router.Get(url, defaultHandler)
	node := router.Route(url, nil)

	if node == nil {
		t.Error("router path is not correct")
	}
}

func TestGraft(t *testing.T) {
	router := NewRouter()
	sub := NewRouter()

	sub.Get("/path/sub", defaultHandler)
	router.Graft("/s", sub)
	node := router.Route("/s/path/sub", nil)

	if node == nil {
		t.Error("router graft is not correct")
	}
}