package event

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Envelope struct {
	Type     string
	ThreadID string
	TurnID   string
	Cwd      string
	Client   string
	Raw      map[string]any
	RawJSON  []byte
}

func Parse(raw []byte) (Envelope, error) {
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return Envelope{}, fmt.Errorf("invalid JSON: %w", err)
	}

	return Envelope{
		Type:     getString(body, "type"),
		ThreadID: getString(body, "thread-id"),
		TurnID:   getString(body, "turn-id"),
		Cwd:      getString(body, "cwd"),
		Client:   getString(body, "client"),
		Raw:      body,
		RawJSON:  append([]byte(nil), raw...),
	}, nil
}

func GetString(raw map[string]any, keys ...string) string {
	for _, key := range keys {
		value := getString(raw, key)
		if value != "" {
			return value
		}
	}
	return ""
}

func GetStringSlice(raw map[string]any, key string) []string {
	value, ok := raw[key]
	if !ok {
		return nil
	}
	items, ok := value.([]any)
	if !ok {
		return nil
	}

	result := make([]string, 0, len(items))
	for _, item := range items {
		text, ok := item.(string)
		if ok && strings.TrimSpace(text) != "" {
			result = append(result, text)
		}
	}
	return result
}

func getString(raw map[string]any, key string) string {
	value, ok := raw[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
