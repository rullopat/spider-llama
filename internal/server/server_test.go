package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rullopat/spider-llama/internal/config"
	"github.com/rullopat/spider-llama/internal/provider"
	"github.com/rullopat/spider-llama/internal/router"
)

func TestChatCompletionsRoutesAliasAndRewritesBackendModel(t *testing.T) {
	var upstreamModel string
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Model string `json:"model"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode upstream body: %v", err)
		}
		upstreamModel = body.Model
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"x","choices":[{"message":{"role":"assistant","content":"ok"}}]}`))
	}))
	defer upstream.Close()

	cfg := config.Config{
		Auth: config.AuthConfig{BearerToken: "token"},
		Nodes: []config.NodeConfig{{
			ID: "local", Provider: "llamacpp", BaseURL: upstream.URL, MaxConcurrency: 1, TimeoutSecs: 5,
		}},
		Models: []config.ModelConfig{{
			ID: "public-model", BackendModel: "backend-model", Node: "local", Aliases: []string{"light-text"}, Capabilities: []string{"text"}, Priority: 1,
		}},
	}
	cfg.ApplyDefaults()
	app := New(cfg, router.NewRegistry(cfg), provider.NewRegistry())

	request := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", strings.NewReader(`{
		"model":"alias:light-text",
		"task":"internal-routing-field",
		"requirements":{"capabilities":["text"]},
		"messages":[{"role":"user","content":"hi"}]
	}`))
	request.Header.Set("Authorization", "Bearer token")
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	app.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", recorder.Code, recorder.Body.String())
	}
	if upstreamModel != "backend-model" {
		t.Fatalf("expected backend model rewrite, got %q", upstreamModel)
	}
	if recorder.Header().Get("X-Spider-Llama-Node") != "local" {
		t.Fatalf("missing selected node header")
	}
}

func TestAuthRequired(t *testing.T) {
	cfg := config.Config{
		Auth:   config.AuthConfig{BearerToken: "token"},
		Nodes:  []config.NodeConfig{{ID: "local", Provider: "llamacpp", BaseURL: "http://127.0.0.1:8080"}},
		Models: []config.ModelConfig{{ID: "model", Node: "local", Capabilities: []string{"text"}}},
	}
	cfg.ApplyDefaults()
	app := New(cfg, router.NewRegistry(cfg), provider.NewRegistry())

	request := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	recorder := httptest.NewRecorder()
	app.Routes().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthorized, got %d", recorder.Code)
	}
}
