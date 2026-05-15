package event

import (
	"encoding/json"
	"fmt"
	"sort"
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
		Type:     GetString(body, "type", "event_type", "eventType", "hook_event_name", "hookEventName"),
		ThreadID: GetString(body, "thread-id", "thread_id", "threadId", "session_id", "sessionId"),
		TurnID:   GetString(body, "turn-id", "turn_id", "turnId"),
		Cwd:      GetString(body, "cwd", "current_working_directory", "currentWorkingDirectory"),
		Client:   GetString(body, "client"),
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

func GetStringSlice(raw map[string]any, keys ...string) []string {
	for _, key := range keys {
		value, ok := raw[key]
		if !ok {
			continue
		}

		switch items := value.(type) {
		case []any:
			result := make([]string, 0, len(items))
			for _, item := range items {
				text, ok := item.(string)
				if ok && isMeaningfulString(text) {
					result = append(result, strings.TrimSpace(text))
				}
			}
			return result
		case []string:
			result := make([]string, 0, len(items))
			for _, text := range items {
				if isMeaningfulString(text) {
					result = append(result, strings.TrimSpace(text))
				}
			}
			return result
		case string:
			if isMeaningfulString(items) {
				return []string{strings.TrimSpace(items)}
			}
		}
	}

	return nil
}

func Keys(raw map[string]any) []string {
	keys := make([]string, 0, len(raw))
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
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
	if !isMeaningfulString(text) {
		return ""
	}
	return strings.TrimSpace(text)
}

func isMeaningfulString(text string) bool {
	value := strings.TrimSpace(text)
	switch strings.ToLower(value) {
	case "", "undefined", "null":
		return false
	default:
		return true
	}
}
