package media

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"MrRSS/internal/database"
	"MrRSS/internal/handlers/core"
)

func TestHandleMediaCacheInfoAndCleanup(t *testing.T) {
	tmp := t.TempDir()
	// Ensure data dir resolves to temp dir
	_ = os.Setenv("APPDATA", tmp)
	_ = os.Setenv("HOME", tmp)
	_ = os.Setenv("XDG_DATA_HOME", tmp)

	db, err := database.NewDB(":memory:")
	if err != nil {
		t.Fatalf("NewDB failed: %v", err)
	}
	if err := db.Init(); err != nil {
		t.Fatalf("db Init failed: %v", err)
	}

	h := core.NewHandler(db, nil, nil)
	// Enable media cache and set small thresholds
	if err := h.DB.SetSetting("media_cache_enabled", "true"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	if err := h.DB.SetSetting("media_cache_max_age_days", "1"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}
	if err := h.DB.SetSetting("media_cache_max_size_mb", "1"); err != nil {
		t.Fatalf("SetSetting failed: %v", err)
	}

	// Call info (GET)
	req := httptest.NewRequest(http.MethodGet, "/media/info", nil)
	rr := httptest.NewRecorder()
	HandleMediaCacheInfo(h, rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 for info, got %d", rr.Code)
	}
	var info map[string]float64
	if err := json.NewDecoder(rr.Body).Decode(&info); err != nil {
		t.Fatalf("decode info failed: %v", err)
	}

	if _, ok := info["cache_size_mb"]; !ok {
		t.Fatalf("info missing cache_size_mb")
	}

	// Call cleanup (POST)
	req2 := httptest.NewRequest(http.MethodPost, "/media/cleanup", nil)
	rr2 := httptest.NewRecorder()
	HandleMediaCacheCleanup(h, rr2, req2)
	if rr2.Code != http.StatusOK {
		t.Fatalf("expected 200 for cleanup, got %d", rr2.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr2.Body).Decode(&resp); err != nil {
		t.Fatalf("decode cleanup failed: %v", err)
	}
	if success, ok := resp["success"].(bool); !ok || !success {
		t.Fatalf("expected cleanup success true, got %v", resp)
	}
}
