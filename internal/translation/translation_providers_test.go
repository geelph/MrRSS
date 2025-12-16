package translation

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type rtFunc func(*http.Request) (*http.Response, error)

func (f rtFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDeepLTranslate_SuccessAndEmpty(t *testing.T) {
	t1 := NewDeepLTranslator("apikey")

	// empty input should return empty without error
	out, err := t1.Translate("", "es")
	if err != nil || out != "" {
		t.Fatalf("expected empty translate for empty input, got %q err=%v", out, err)
	}

	// Mock client response for success
	t1.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"translations":[{"text":"Hola"}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	out2, err := t1.Translate("Hello", "es")
	if err != nil {
		t.Fatalf("DeepL translate failed: %v", err)
	}
	if out2 != "Hola" {
		t.Fatalf("expected Hola, got %s", out2)
	}
}

func TestBaiduTranslate_SuccessAndEmpty(t *testing.T) {
	t1 := NewBaiduTranslator("appid", "secret")

	out, err := t1.Translate("", "en")
	if err != nil || out != "" {
		t.Fatalf("expected empty translate for empty input, got %q err=%v", out, err)
	}

	// Mock baidu response
	t1.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"trans_result":[{"src":"Hello","dst":"你好"}], "error_code":"52000"}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	out2, err := t1.Translate("Hello", "zh")
	if err != nil {
		t.Fatalf("Baidu translate failed: %v", err)
	}
	if out2 != "你好" {
		t.Fatalf("expected 你好, got %s", out2)
	}
}

func TestAITranslate_SuccessAndEmpty(t *testing.T) {
	t1 := NewAITranslator("apikey", "https://api.test", "m1")

	out, err := t1.Translate("", "en")
	if err != nil || out != "" {
		t.Fatalf("expected empty translate for empty input, got %q err=%v", out, err)
	}

	// Mock AI response
	t1.client = &http.Client{Transport: rtFunc(func(req *http.Request) (*http.Response, error) {
		body := `{"choices":[{"message":{"content":"Bonjour"}}]}`
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: http.Header{"Content-Type": {"application/json"}}}, nil
	}), Timeout: 5 * time.Second}

	out2, err := t1.Translate("Hello", "fr")
	if err != nil {
		t.Fatalf("AI translate failed: %v", err)
	}
	if out2 != "Bonjour" {
		t.Fatalf("expected Bonjour, got %s", out2)
	}
}
