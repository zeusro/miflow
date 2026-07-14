package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestChat(t *testing.T) {
	wantContent := `{"action":"turn_on","device_name":"客厅灯"}`
	 srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/chat" {
			t.Errorf("path = %s, want /api/chat", r.URL.Path)
		}
		var req chatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Model != "test-model" {
			t.Errorf("model = %s, want test-model", req.Model)
		}
		if req.Format != "json" {
			t.Errorf("format = %s, want json", req.Format)
		}
		resp := chatResponse{
			Message: chatMessage{Role: "assistant", Content: wantContent},
			Done:    true,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-model", 0)
	got, err := c.Chat(context.Background(), "system", "打开客厅灯")
	if err != nil {
		t.Fatalf("Chat error: %v", err)
	}
	if got != wantContent {
		t.Errorf("Chat = %q, want %q", got, wantContent)
	}
}

func TestChatHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "model not found", http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "missing", 0)
	_, err := c.Chat(context.Background(), "", "hello")
	if err == nil {
		t.Fatal("expected error for http 404")
	}
}

func TestNewClientDefaults(t *testing.T) {
	c := NewClient("", "", 0)
	if c.Host != "http://localhost:11434" {
		t.Errorf("host = %s, want http://localhost:11434", c.Host)
	}
	if c.Model != "qwen2.5" {
		t.Errorf("model = %s, want qwen2.5", c.Model)
	}
	if c.Timeout <= 0 {
		t.Error("timeout should be > 0")
	}
}
