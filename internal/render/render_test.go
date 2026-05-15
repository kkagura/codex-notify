package render

import (
	"strings"
	"testing"

	"codex-notify/internal/event"
)

func TestBuildRendersPermissionRequest(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PermissionRequest",
		"session_id": "session-1",
		"turn_id": "turn-1",
		"cwd": "E:\\wang\\codex-notify",
		"model": "gpt-5.5",
		"permission_mode": "default",
		"tool_name": "Bash",
		"tool_input": {
			"command": "go test ./...",
			"description": "Run tests"
		}
	}`)

	env, err := event.Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	notification := Build(env)
	if !strings.Contains(notification.Title, "Permission Request") {
		t.Fatalf("Title = %q", notification.Title)
	}
	for _, want := range []string{"Tool: Bash", "Command: go test ./...", "Reason: Run tests", "Mode: default", "Model: gpt-5.5"} {
		if !strings.Contains(notification.Message, want) {
			t.Fatalf("Message = %q, want %q", notification.Message, want)
		}
	}
	if strings.Contains(notification.Title, "unknown") || strings.Contains(notification.Message, "undefined") {
		t.Fatalf("unexpected fallback text in notification: title=%q message=%q", notification.Title, notification.Message)
	}
}
