package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func parseChatMeta(body map[string]json.RawMessage) (chatMeta, error) {
	var meta chatMeta
	if raw, ok := body["model"]; ok {
		if err := json.Unmarshal(raw, &meta.Model); err != nil {
			return meta, fmt.Errorf("model must be a string")
		}
	}
	if raw, ok := body["task"]; ok {
		if err := json.Unmarshal(raw, &meta.Task); err != nil {
			return meta, fmt.Errorf("task must be a string")
		}
	}
	if raw, ok := body["stream"]; ok {
		if err := json.Unmarshal(raw, &meta.Stream); err != nil {
			return meta, fmt.Errorf("stream must be a boolean")
		}
	}
	if raw, ok := body["requirements"]; ok {
		if err := json.Unmarshal(raw, &meta.Requirements); err != nil {
			return meta, fmt.Errorf("requirements must match the requirements schema")
		}
	}
	return meta, nil
}

func buildBackendPayload(body map[string]json.RawMessage, backendModel string) ([]byte, error) {
	cleaned := make(map[string]json.RawMessage, len(body)+1)
	for key, value := range body {
		switch key {
		case "requirements", "task":
			continue
		default:
			cleaned[key] = value
		}
	}
	modelBytes, err := json.Marshal(backendModel)
	if err != nil {
		return nil, err
	}
	cleaned["model"] = modelBytes
	return json.Marshal(cleaned)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{
		"error": map[string]any{
			"message": err.Error(),
		},
	})
}
