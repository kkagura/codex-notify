package event

import "testing"

func TestParseAcceptsCommonFieldNameVariants(t *testing.T) {
	raw := []byte(`{
		"eventType": "agent-turn-complete",
		"thread_id": "thread-1",
		"turnId": "turn-1",
		"currentWorkingDirectory": "E:\\wang\\codex-notify",
		"client": "codex-tui"
	}`)

	env, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if env.Type != "agent-turn-complete" {
		t.Fatalf("Type = %q", env.Type)
	}
	if env.ThreadID != "thread-1" {
		t.Fatalf("ThreadID = %q", env.ThreadID)
	}
	if env.TurnID != "turn-1" {
		t.Fatalf("TurnID = %q", env.TurnID)
	}
	if env.Cwd != `E:\wang\codex-notify` {
		t.Fatalf("Cwd = %q", env.Cwd)
	}
}

func TestParseAcceptsPermissionRequestHookFields(t *testing.T) {
	raw := []byte(`{
		"hook_event_name": "PermissionRequest",
		"session_id": "session-1",
		"turn_id": "turn-1",
		"cwd": "E:\\wang\\codex-notify",
		"tool_name": "Bash"
	}`)

	env, err := Parse(raw)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if env.Type != "PermissionRequest" {
		t.Fatalf("Type = %q", env.Type)
	}
	if env.ThreadID != "session-1" {
		t.Fatalf("ThreadID = %q", env.ThreadID)
	}
	if env.TurnID != "turn-1" {
		t.Fatalf("TurnID = %q", env.TurnID)
	}
}

func TestGetStringIgnoresUndefinedAndNullStrings(t *testing.T) {
	raw := map[string]any{
		"primary":  "undefined",
		"fallback": "value",
		"nullish":  "null",
	}

	if got := GetString(raw, "primary", "fallback"); got != "value" {
		t.Fatalf("GetString() = %q", got)
	}
	if got := GetString(raw, "nullish"); got != "" {
		t.Fatalf("GetString(nullish) = %q", got)
	}
}

func TestGetStringSliceAcceptsVariantsAndFiltersUndefined(t *testing.T) {
	raw := map[string]any{
		"inputMessages": []any{"undefined", "first prompt"},
	}

	got := GetStringSlice(raw, "input-messages", "inputMessages")
	if len(got) != 1 || got[0] != "first prompt" {
		t.Fatalf("GetStringSlice() = %#v", got)
	}
}
