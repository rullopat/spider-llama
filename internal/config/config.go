package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	DefaultListen                = ":8088"
	DefaultMaxBodyBytes          = int64(2 * 1024 * 1024)
	DefaultRequestTimeoutSeconds = 180
	DefaultNodeTimeoutSeconds    = 180
)

type Config struct {
	Listen        string              `json:"listen"`
	Auth          AuthConfig          `json:"auth"`
	RequestLimits RequestLimitsConfig `json:"request_limits"`
	Nodes         []NodeConfig        `json:"nodes"`
	Models        []ModelConfig       `json:"models"`
	Routes        map[string]Route    `json:"routes"`
}

type AuthConfig struct {
	BearerToken    string `json:"bearer_token"`
	BearerTokenEnv string `json:"bearer_token_env"`
}

type RequestLimitsConfig struct {
	MaxBodyBytes       int64 `json:"max_body_bytes"`
	RequestTimeoutSecs int   `json:"request_timeout_seconds"`
}

type NodeConfig struct {
	ID             string   `json:"id"`
	Provider       string   `json:"provider"`
	BaseURL        string   `json:"base_url"`
	Tags           []string `json:"tags"`
	MaxConcurrency int      `json:"max_concurrency"`
	TimeoutSecs    int      `json:"timeout_seconds"`
	APIKey         string   `json:"api_key"`
	APIKeyEnv      string   `json:"api_key_env"`
	Disabled       bool     `json:"disabled"`
}

type ModelConfig struct {
	ID            string   `json:"id"`
	BackendModel  string   `json:"backend_model"`
	Node          string   `json:"node"`
	Aliases       []string `json:"aliases"`
	Capabilities  []string `json:"capabilities"`
	Tags          []string `json:"tags"`
	ContextTokens int      `json:"context_tokens"`
	Priority      int      `json:"priority"`
	Disabled      bool     `json:"disabled"`
}

type Route struct {
	Require Requirements `json:"require"`
	Prefer  Preferences  `json:"prefer"`
}

type Requirements struct {
	Capabilities     []string `json:"capabilities"`
	Tags             []string `json:"tags"`
	NodeTags         []string `json:"node_tags"`
	MinContextTokens int      `json:"min_context_tokens"`
}

type Preferences struct {
	Tags     []string `json:"tags"`
	NodeTags []string `json:"node_tags"`
}

func Load(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}

	cfg.ApplyDefaults()
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) ApplyDefaults() {
	if strings.TrimSpace(c.Listen) == "" {
		c.Listen = DefaultListen
	}
	if c.RequestLimits.MaxBodyBytes <= 0 {
		c.RequestLimits.MaxBodyBytes = DefaultMaxBodyBytes
	}
	if c.RequestLimits.RequestTimeoutSecs <= 0 {
		c.RequestLimits.RequestTimeoutSecs = DefaultRequestTimeoutSeconds
	}
	for i := range c.Nodes {
		if c.Nodes[i].MaxConcurrency <= 0 {
			c.Nodes[i].MaxConcurrency = 1
		}
		if c.Nodes[i].TimeoutSecs <= 0 {
			c.Nodes[i].TimeoutSecs = DefaultNodeTimeoutSeconds
		}
		c.Nodes[i].Provider = normalize(c.Nodes[i].Provider)
		c.Nodes[i].Tags = normalizeSlice(c.Nodes[i].Tags)
	}
	for i := range c.Models {
		if c.Models[i].BackendModel == "" {
			c.Models[i].BackendModel = c.Models[i].ID
		}
		c.Models[i].Aliases = normalizeSlice(c.Models[i].Aliases)
		c.Models[i].Capabilities = normalizeSlice(c.Models[i].Capabilities)
		c.Models[i].Tags = normalizeSlice(c.Models[i].Tags)
	}
	if c.Routes == nil {
		c.Routes = map[string]Route{}
	}
	normalizedRoutes := make(map[string]Route, len(c.Routes))
	for name, route := range c.Routes {
		route.Require.Capabilities = normalizeSlice(route.Require.Capabilities)
		route.Require.Tags = normalizeSlice(route.Require.Tags)
		route.Require.NodeTags = normalizeSlice(route.Require.NodeTags)
		route.Prefer.Tags = normalizeSlice(route.Prefer.Tags)
		route.Prefer.NodeTags = normalizeSlice(route.Prefer.NodeTags)
		normalizedRoutes[normalize(name)] = route
	}
	c.Routes = normalizedRoutes
}

func (c Config) Validate() error {
	var problems []string
	if len(c.Nodes) == 0 {
		problems = append(problems, "at least one node is required")
	}
	if len(c.Models) == 0 {
		problems = append(problems, "at least one model is required")
	}

	nodeIDs := map[string]struct{}{}
	for _, node := range c.Nodes {
		if node.ID == "" {
			problems = append(problems, "node id is required")
		}
		if node.Provider == "" {
			problems = append(problems, fmt.Sprintf("node %q provider is required", node.ID))
		}
		if node.BaseURL == "" {
			problems = append(problems, fmt.Sprintf("node %q base_url is required", node.ID))
		}
		if _, exists := nodeIDs[node.ID]; exists {
			problems = append(problems, fmt.Sprintf("duplicate node id %q", node.ID))
		}
		nodeIDs[node.ID] = struct{}{}
	}

	modelIDs := map[string]struct{}{}
	aliases := map[string]string{}
	for _, model := range c.Models {
		if model.ID == "" {
			problems = append(problems, "model id is required")
		}
		if _, exists := modelIDs[model.ID]; exists {
			problems = append(problems, fmt.Sprintf("duplicate model id %q", model.ID))
		}
		modelIDs[model.ID] = struct{}{}

		if _, ok := nodeIDs[model.Node]; !ok {
			problems = append(problems, fmt.Sprintf("model %q references unknown node %q", model.ID, model.Node))
		}

		for _, alias := range model.Aliases {
			if existing, exists := aliases[alias]; exists {
				problems = append(problems, fmt.Sprintf("alias %q is used by both %q and %q", alias, existing, model.ID))
			}
			aliases[alias] = model.ID
		}
	}

	if len(problems) > 0 {
		return errors.New(strings.Join(problems, "; "))
	}
	return nil
}

func (a AuthConfig) EffectiveBearerToken() string {
	if a.BearerTokenEnv != "" {
		if token := os.Getenv(a.BearerTokenEnv); token != "" {
			return token
		}
	}
	return a.BearerToken
}

func (n NodeConfig) EffectiveAPIKey() string {
	if n.APIKeyEnv != "" {
		if token := os.Getenv(n.APIKeyEnv); token != "" {
			return token
		}
	}
	return n.APIKey
}

func normalizeSlice(values []string) []string {
	if len(values) == 0 {
		return values
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = normalize(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
