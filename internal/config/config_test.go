package config

import "testing"

func TestGetString_KnownKeys(t *testing.T) {
	// A few representative keys
	if v := GetString("language"); v == "" {
		t.Fatalf("expected non-empty language default")
	}
	if v := GetString("update_interval"); v == "" {
		t.Fatalf("expected non-empty update_interval default")
	}
	if v := GetString("show_article_preview_images"); v == "" {
		t.Fatalf("expected non-empty show_article_preview_images default")
	}
}

func TestGetString_UnknownKey(t *testing.T) {
	if v := GetString("this_key_does_not_exist"); v != "" {
		t.Fatalf("expected empty string for unknown key, got %q", v)
	}
}
