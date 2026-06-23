package provider

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/rullopat/spider-llama/internal/config"
)

type LlamaCPP struct{}

func (LlamaCPP) Health(ctx context.Context, node config.NodeConfig) Health {
	endpoint, err := joinURL(node.BaseURL, "/health")
	if err != nil {
		return Health{OK: false, Status: "invalid_base_url", Detail: err.Error()}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Health{OK: false, Status: "request_error", Detail: err.Error()}
	}
	addAuth(request, node)

	response, err := clientFor(node).Do(request)
	if err != nil {
		return Health{OK: false, Status: "unreachable", Detail: err.Error()}
	}
	defer response.Body.Close()

	if response.StatusCode >= 200 && response.StatusCode < 300 {
		return Health{OK: true, Status: "ok"}
	}
	return Health{OK: false, Status: fmt.Sprintf("http_%d", response.StatusCode)}
}

func (LlamaCPP) Chat(ctx context.Context, node config.NodeConfig, model config.ModelConfig, body []byte) Response {
	endpoint, err := joinURL(node.BaseURL, "/v1/chat/completions")
	if err != nil {
		return Response{StatusCode: http.StatusBadGateway, Err: err}
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return Response{StatusCode: http.StatusBadGateway, Err: err}
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")
	addAuth(request, node)

	response, err := clientFor(node).Do(request)
	if err != nil {
		return Response{StatusCode: http.StatusBadGateway, Err: err}
	}
	return readResponse(response)
}

func addAuth(request *http.Request, node config.NodeConfig) {
	if apiKey := node.EffectiveAPIKey(); apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}
}

func joinURL(base, path string) (string, error) {
	parsed, err := url.Parse(strings.TrimRight(base, "/"))
	if err != nil {
		return "", err
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return "", fmt.Errorf("base_url must include scheme and host")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/") + path
	return parsed.String(), nil
}
