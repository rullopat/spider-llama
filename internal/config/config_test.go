package config

import "testing"

func TestApplyDefaults(t *testing.T) {
	cfg := Config{
		Nodes: []NodeConfig{{ID: "n1", Provider: "LlamaCPP", BaseURL: "http://127.0.0.1:8080"}},
		Models: []ModelConfig{{
			ID:           "m1",
			Node:         "n1",
			Aliases:      []string{"Light-Text", "light-text"},
			Capabilities: []string{"Text"},
			Tags:         []string{"Light"},
		}},
	}

	cfg.ApplyDefaults()

	if cfg.Listen != DefaultListen {
		t.Fatalf("expected default listen %q, got %q", DefaultListen, cfg.Listen)
	}
	if cfg.Nodes[0].Provider != "llamacpp" {
		t.Fatalf("provider was not normalized: %q", cfg.Nodes[0].Provider)
	}
	if cfg.Nodes[0].MaxConcurrency != 1 {
		t.Fatalf("expected default max concurrency 1, got %d", cfg.Nodes[0].MaxConcurrency)
	}
	if cfg.Models[0].BackendModel != "m1" {
		t.Fatalf("expected backend model to default to id")
	}
	if len(cfg.Models[0].Aliases) != 1 || cfg.Models[0].Aliases[0] != "light-text" {
		t.Fatalf("aliases were not normalized/deduped: %#v", cfg.Models[0].Aliases)
	}
}

func TestValidateRejectsUnknownNode(t *testing.T) {
	cfg := Config{
		Nodes:  []NodeConfig{{ID: "n1", Provider: "llamacpp", BaseURL: "http://127.0.0.1:8080"}},
		Models: []ModelConfig{{ID: "m1", Node: "missing"}},
	}
	cfg.ApplyDefaults()

	if err := cfg.Validate(); err == nil {
		t.Fatalf("expected validation error")
	}
}
