package provider

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rullopat/spider-llama/internal/config"
)

func TestLlamaCPPChatPostsToOpenAICompatibleEndpoint(t *testing.T) {
	var received map[string]any
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret" {
			t.Fatalf("missing backend auth header")
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode upstream request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"chatcmpl-test","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer upstream.Close()

	node := config.NodeConfig{BaseURL: upstream.URL, TimeoutSecs: 5, APIKey: "secret"}
	model := config.ModelConfig{ID: "alias-model", BackendModel: "backend-model"}
	body := []byte(`{"model":"alias-model","messages":[{"role":"user","content":"hi"}]}`)

	response := LlamaCPP{}.Chat(t.Context(), node, model, body)
	if response.Err != nil {
		t.Fatalf("chat failed: %v", response.Err)
	}
	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}
	if received["model"] != "alias-model" {
		t.Fatalf("provider should receive prepared payload unchanged, got model %#v", received["model"])
	}
}

func TestLlamaCPPHealth(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

	health := LlamaCPP{}.Health(t.Context(), config.NodeConfig{BaseURL: upstream.URL, TimeoutSecs: 5})
	if !health.OK {
		t.Fatalf("expected health OK, got %#v", health)
	}
}
