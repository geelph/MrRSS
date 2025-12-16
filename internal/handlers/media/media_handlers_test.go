package media

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"MrRSS/internal/database"
	corepkg "MrRSS/internal/handlers/core"
)

func setupHandler(t *testing.T) *corepkg.Handler {
	t.Helper()
	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB error: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init error: %v", err)
	}
	return corepkg.NewHandler(db, nil, nil)
}

func TestHandleMediaProxy_MethodNotAllowed(t *testing.T) {
	h := setupHandler(t)
	req := httptest.NewRequest(http.MethodPost, "/media/proxy", nil)
	rr := httptest.NewRecorder()

	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected %d got %d", http.StatusMethodNotAllowed, rr.Code)
	}
}

func TestHandleMediaProxy_CacheDisabled(t *testing.T) {
	h := setupHandler(t)
	req := httptest.NewRequest(http.MethodGet, "/media/proxy", nil)
	rr := httptest.NewRecorder()

	// By default media_cache_enabled is not true
	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected %d got %d", http.StatusForbidden, rr.Code)
	}
}

func TestHandleMediaProxy_MissingURL(t *testing.T) {
	h := setupHandler(t)
	// enable cache setting
	_ = h.DB.SetSetting("media_cache_enabled", "true")

	req := httptest.NewRequest(http.MethodGet, "/media/proxy", nil)
	rr := httptest.NewRecorder()

	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}
}

func TestHandleMediaProxy_InvalidURL(t *testing.T) {
	h := setupHandler(t)
	_ = h.DB.SetSetting("media_cache_enabled", "true")

	req := httptest.NewRequest(http.MethodGet, "/media/proxy?url=ftp://example.com/file.jpg", nil)
	rr := httptest.NewRecorder()

	HandleMediaProxy(h, rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected %d got %d", http.StatusBadRequest, rr.Code)
	}
}
