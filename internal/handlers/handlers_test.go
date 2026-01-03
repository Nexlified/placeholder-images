package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/golang-lru/v2"

	"go-avatars/internal/config"
	"go-avatars/internal/render"
)

func TestAvatarHandlerDefaults(t *testing.T) {
	renderer, err := render.New()
	if err != nil {
		t.Fatalf("renderer init: %v", err)
	}
	cache, _ := lru.New[string, []byte](1)
	svc := NewService(renderer, cache, config.DefaultServerConfig())
	mux := http.NewServeMux()
	svc.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/avatar/", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Fatalf("expected content-type image/png got %s", ct)
	}
	if rec.Body.Len() == 0 {
		t.Fatal("expected body to contain png data")
	}
}
