package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/rullopat/spider-llama/internal/config"
	"github.com/rullopat/spider-llama/internal/provider"
	"github.com/rullopat/spider-llama/internal/router"
)

type Server struct {
	cfg       config.Config
	registry  *router.Registry
	providers *provider.Registry
	token     string
}

type chatMeta struct {
	Model        string              `json:"model"`
	Task         string              `json:"task"`
	Stream       bool                `json:"stream"`
	Requirements config.Requirements `json:"requirements"`
}

func New(cfg config.Config, registry *router.Registry, providers *provider.Registry) *Server {
	return &Server{
		cfg:       cfg,
		registry:  registry,
		providers: providers,
		token:     cfg.Auth.EffectiveBearerToken(),
	}
}

func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", s.health)
	mux.HandleFunc("GET /v1/nodes", s.requireAuth(s.nodes))
	mux.HandleFunc("GET /v1/models", s.requireAuth(s.models))
	mux.HandleFunc("POST /v1/route", s.requireAuth(s.route))
	mux.HandleFunc("POST /v1/chat/completions", s.requireAuth(s.chatCompletions))
	return mux
}

func (s *Server) health(w http.ResponseWriter, r *http.Request) {
	type nodeHealth struct {
		ID       string          `json:"id"`
		Provider string          `json:"provider"`
		Load     loadInfo        `json:"load"`
		Health   provider.Health `json:"health"`
	}

	nodes := s.registry.Nodes()
	results := make([]nodeHealth, len(nodes))
	var wg sync.WaitGroup
	for i, node := range nodes {
		i, node := i, node
		wg.Add(1)
		go func() {
			defer wg.Done()
			current, max := s.registry.NodeLoad(node.ID)
			backend, ok := s.providers.Get(node.Provider)
			if !ok {
				results[i] = nodeHealth{
					ID:       node.ID,
					Provider: node.Provider,
					Load:     loadInfo{Current: current, Max: max},
					Health:   provider.Health{OK: false, Status: "unknown_provider"},
				}
				return
			}
			ctx, cancel := context.WithTimeout(r.Context(), time.Duration(node.TimeoutSecs)*time.Second)
			defer cancel()
			results[i] = nodeHealth{
				ID:       node.ID,
				Provider: node.Provider,
				Load:     loadInfo{Current: current, Max: max},
				Health:   backend.Health(ctx, node),
			}
		}()
	}
	wg.Wait()

	ok := true
	for _, result := range results {
		if !result.Health.OK {
			ok = false
			break
		}
	}

	status := http.StatusOK
	if !ok {
		status = http.StatusServiceUnavailable
	}
	writeJSON(w, status, map[string]any{
		"ok":    ok,
		"nodes": results,
	})
}

func (s *Server) nodes(w http.ResponseWriter, r *http.Request) {
	type nodeView struct {
		ID       string   `json:"id"`
		Provider string   `json:"provider"`
		BaseURL  string   `json:"base_url"`
		Tags     []string `json:"tags"`
		Load     loadInfo `json:"load"`
		Disabled bool     `json:"disabled"`
	}

	nodes := s.registry.Nodes()
	views := make([]nodeView, 0, len(nodes))
	for _, node := range nodes {
		current, max := s.registry.NodeLoad(node.ID)
		views = append(views, nodeView{
			ID:       node.ID,
			Provider: node.Provider,
			BaseURL:  node.BaseURL,
			Tags:     node.Tags,
			Load:     loadInfo{Current: current, Max: max},
			Disabled: node.Disabled,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"data": views})
}

func (s *Server) models(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"data": s.registry.Models()})
}

func (s *Server) route(w http.ResponseWriter, r *http.Request) {
	body, err := s.readJSONBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	meta, err := parseChatMeta(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	candidates, err := s.registry.Select(router.SelectionRequest{
		Model:        meta.Model,
		Task:         meta.Task,
		Requirements: meta.Requirements,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"data": candidates})
}

func (s *Server) chatCompletions(w http.ResponseWriter, r *http.Request) {
	body, err := s.readJSONBody(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	meta, err := parseChatMeta(body)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if meta.Stream {
		writeError(w, http.StatusNotImplemented, errors.New("streaming responses are not supported in the MVP"))
		return
	}

	candidates, err := s.registry.Select(router.SelectionRequest{
		Model:        meta.Model,
		Task:         meta.Task,
		Requirements: meta.Requirements,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}

	var lastErr error
	for _, candidate := range candidates {
		release, ok := s.registry.TryAcquire(candidate.Node.ID)
		if !ok {
			lastErr = fmt.Errorf("node %q is at max concurrency", candidate.Node.ID)
			continue
		}

		response := s.callBackend(r.Context(), candidate, body)
		release()

		if response.Err != nil {
			lastErr = response.Err
			continue
		}

		w.Header().Set("Content-Type", response.ContentType)
		w.Header().Set("X-Spider-Llama-Node", candidate.Node.ID)
		w.Header().Set("X-Spider-Llama-Model", candidate.Model.ID)
		w.WriteHeader(response.StatusCode)
		_, _ = w.Write(response.Body)
		return
	}

	if lastErr == nil {
		lastErr = errors.New("no available backend node")
	}
	writeError(w, http.StatusServiceUnavailable, lastErr)
}

func (s *Server) callBackend(ctx context.Context, candidate router.Candidate, body map[string]json.RawMessage) provider.Response {
	backend, ok := s.providers.Get(candidate.Node.Provider)
	if !ok {
		return provider.Response{StatusCode: http.StatusBadGateway, Err: fmt.Errorf("unknown provider %q", candidate.Node.Provider)}
	}

	payload, err := buildBackendPayload(body, candidate.Model.BackendModel)
	if err != nil {
		return provider.Response{StatusCode: http.StatusBadRequest, Err: err}
	}

	timeout := time.Duration(candidate.Node.TimeoutSecs) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return backend.Chat(ctx, candidate.Node, candidate.Model, payload)
}

func (s *Server) readJSONBody(r *http.Request) (map[string]json.RawMessage, error) {
	reader := http.MaxBytesReader(nil, r.Body, s.cfg.RequestLimits.MaxBodyBytes)
	defer r.Body.Close()

	decoder := json.NewDecoder(reader)
	decoder.DisallowUnknownFields()
	var body map[string]json.RawMessage
	if err := decoder.Decode(&body); err != nil {
		return nil, err
	}
	return body, nil
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if s.token == "" {
			next(w, r)
			return
		}

		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") || strings.TrimPrefix(header, "Bearer ") != s.token {
			writeError(w, http.StatusUnauthorized, errors.New("missing or invalid bearer token"))
			return
		}
		next(w, r)
	}
}

type loadInfo struct {
	Current int `json:"current"`
	Max     int `json:"max"`
}
