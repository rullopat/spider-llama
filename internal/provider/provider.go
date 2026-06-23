package provider

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rullopat/spider-llama/internal/config"
)

type Registry struct {
	backends map[string]Backend
}

type Backend interface {
	Health(ctx context.Context, node config.NodeConfig) Health
	Chat(ctx context.Context, node config.NodeConfig, model config.ModelConfig, body []byte) Response
}

type Health struct {
	OK     bool   `json:"ok"`
	Status string `json:"status"`
	Detail string `json:"detail,omitempty"`
}

type Response struct {
	StatusCode  int
	ContentType string
	Body        []byte
	Err         error
}

func NewRegistry() *Registry {
	return &Registry{
		backends: map[string]Backend{
			"llamacpp":  LlamaCPP{},
			"llama.cpp": LlamaCPP{},
		},
	}
}

func (r *Registry) Get(name string) (Backend, bool) {
	backend, ok := r.backends[strings.ToLower(strings.TrimSpace(name))]
	return backend, ok
}

func clientFor(node config.NodeConfig) *http.Client {
	return &http.Client{Timeout: time.Duration(node.TimeoutSecs) * time.Second}
}

func readResponse(response *http.Response) Response {
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return Response{StatusCode: http.StatusBadGateway, Err: err}
	}
	contentType := response.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/json"
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return Response{
			StatusCode:  response.StatusCode,
			ContentType: contentType,
			Body:        body,
			Err:         fmt.Errorf("backend returned HTTP %d", response.StatusCode),
		}
	}
	return Response{StatusCode: response.StatusCode, ContentType: contentType, Body: body}
}
