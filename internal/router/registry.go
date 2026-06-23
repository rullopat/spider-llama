package router

import (
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/rullopat/spider-llama/internal/config"
)

type Registry struct {
	cfg          config.Config
	nodesByID    map[string]config.NodeConfig
	modelsByID   map[string]config.ModelConfig
	modelByAlias map[string]string
	nodeStates   map[string]*NodeState
	mu           sync.RWMutex
}

type NodeState struct {
	node config.NodeConfig
	sem  chan struct{}
}

type SelectionRequest struct {
	Model        string
	Task         string
	Requirements config.Requirements
}

type Candidate struct {
	Model config.ModelConfig `json:"model"`
	Node  config.NodeConfig  `json:"node"`
	Score int                `json:"score"`
}

func NewRegistry(cfg config.Config) *Registry {
	nodesByID := make(map[string]config.NodeConfig, len(cfg.Nodes))
	nodeStates := make(map[string]*NodeState, len(cfg.Nodes))
	for _, node := range cfg.Nodes {
		nodesByID[node.ID] = node
		nodeStates[node.ID] = &NodeState{
			node: node,
			sem:  make(chan struct{}, node.MaxConcurrency),
		}
	}

	modelsByID := make(map[string]config.ModelConfig, len(cfg.Models))
	modelByAlias := map[string]string{}
	for _, model := range cfg.Models {
		modelsByID[model.ID] = model
		for _, alias := range model.Aliases {
			modelByAlias[alias] = model.ID
		}
	}

	return &Registry{
		cfg:          cfg,
		nodesByID:    nodesByID,
		modelsByID:   modelsByID,
		modelByAlias: modelByAlias,
		nodeStates:   nodeStates,
	}
}

func (r *Registry) Select(req SelectionRequest) ([]Candidate, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	modelName := normalizeModelName(req.Model)
	if modelName != "" && modelName != "auto" {
		modelID, ok := r.resolveModelID(modelName)
		if !ok {
			return nil, fmt.Errorf("unknown model or alias %q", req.Model)
		}
		model := r.modelsByID[modelID]
		node, ok := r.nodesByID[model.Node]
		if !ok || model.Disabled || node.Disabled {
			return nil, fmt.Errorf("model %q is unavailable", req.Model)
		}
		return []Candidate{{Model: model, Node: node, Score: model.Priority}}, nil
	}

	route, hasRoute := r.cfg.Routes[normalize(req.Task)]
	if !hasRoute {
		route, hasRoute = r.cfg.Routes["default"]
	}

	requirements := mergeRequirements(route.Require, req.Requirements)
	preferences := route.Prefer

	var candidates []Candidate
	for _, model := range r.cfg.Models {
		if model.Disabled {
			continue
		}
		node, ok := r.nodesByID[model.Node]
		if !ok || node.Disabled {
			continue
		}
		if !matchesRequirements(model, node, requirements) {
			continue
		}
		score := scoreCandidate(model, node, preferences)
		candidates = append(candidates, Candidate{Model: model, Node: node, Score: score})
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Score == candidates[j].Score {
			return candidates[i].Model.ID < candidates[j].Model.ID
		}
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) == 0 {
		return nil, fmt.Errorf("no model matches requested requirements")
	}
	return candidates, nil
}

func (r *Registry) TryAcquire(nodeID string) (func(), bool) {
	r.mu.RLock()
	state := r.nodeStates[nodeID]
	r.mu.RUnlock()
	if state == nil {
		return nil, false
	}

	select {
	case state.sem <- struct{}{}:
		return func() { <-state.sem }, true
	default:
		return nil, false
	}
}

func (r *Registry) NodeLoad(nodeID string) (current int, max int) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	state := r.nodeStates[nodeID]
	if state == nil {
		return 0, 0
	}
	return len(state.sem), cap(state.sem)
}

func (r *Registry) Nodes() []config.NodeConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	nodes := append([]config.NodeConfig(nil), r.cfg.Nodes...)
	sort.Slice(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	return nodes
}

func (r *Registry) Models() []config.ModelConfig {
	r.mu.RLock()
	defer r.mu.RUnlock()
	models := append([]config.ModelConfig(nil), r.cfg.Models...)
	sort.Slice(models, func(i, j int) bool { return models[i].ID < models[j].ID })
	return models
}

func (r *Registry) resolveModelID(name string) (string, bool) {
	if strings.HasPrefix(name, "alias:") {
		name = strings.TrimPrefix(name, "alias:")
	}
	if _, ok := r.modelsByID[name]; ok {
		return name, true
	}
	modelID, ok := r.modelByAlias[name]
	return modelID, ok
}

func matchesRequirements(model config.ModelConfig, node config.NodeConfig, req config.Requirements) bool {
	if req.MinContextTokens > 0 && model.ContextTokens > 0 && model.ContextTokens < req.MinContextTokens {
		return false
	}
	if !containsAll(model.Capabilities, req.Capabilities) {
		return false
	}
	if !containsAll(model.Tags, req.Tags) {
		return false
	}
	if !containsAll(node.Tags, req.NodeTags) {
		return false
	}
	return true
}

func scoreCandidate(model config.ModelConfig, node config.NodeConfig, pref config.Preferences) int {
	score := model.Priority
	score += 10 * overlapCount(model.Tags, pref.Tags)
	score += 5 * overlapCount(node.Tags, pref.NodeTags)
	return score
}

func mergeRequirements(base, override config.Requirements) config.Requirements {
	return config.Requirements{
		Capabilities:     appendUnique(base.Capabilities, override.Capabilities),
		Tags:             appendUnique(base.Tags, override.Tags),
		NodeTags:         appendUnique(base.NodeTags, override.NodeTags),
		MinContextTokens: maxInt(base.MinContextTokens, override.MinContextTokens),
	}
}

func containsAll(values []string, required []string) bool {
	if len(required) == 0 {
		return true
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	for _, value := range required {
		if _, ok := set[value]; !ok {
			return false
		}
	}
	return true
}

func overlapCount(values []string, preferred []string) int {
	if len(values) == 0 || len(preferred) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		set[value] = struct{}{}
	}
	count := 0
	for _, value := range preferred {
		if _, ok := set[value]; ok {
			count++
		}
	}
	return count
}

func appendUnique(a, b []string) []string {
	if len(a) == 0 {
		return append([]string(nil), b...)
	}
	out := append([]string(nil), a...)
	seen := make(map[string]struct{}, len(out)+len(b))
	for _, value := range out {
		seen[value] = struct{}{}
	}
	for _, value := range b {
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

func normalizeModelName(value string) string {
	return strings.TrimSpace(value)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
