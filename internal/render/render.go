package render

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"codex-notify/internal/event"
	"codex-notify/internal/notify"
)

const (
	defaultAppID        = "codex-notify"
	defaultMaxBodyRunes = 220
)

func Build(env event.Envelope) notify.Notification {
	switch env.Type {
	case "agent-turn-complete":
		return renderAgentTurnComplete(env)
	case "PermissionRequest":
		return renderPermissionRequest(env)
	default:
		return renderDefault(env)
	}
}

func renderAgentTurnComplete(env event.Envelope) notify.Notification {
	inputs := event.GetStringSlice(env.Raw, "input-messages", "input_messages", "inputMessages")
	project := projectName(env.Cwd)

	parts := make([]string, 0, 2)
	if len(inputs) > 0 {
		parts = append(parts, cleanText(inputs[0], 90))
	}

	assistant := event.GetString(env.Raw, "last-assistant-message", "last_assistant_message", "lastAssistantMessage")
	if assistant != "" {
		parts = append(parts, cleanText(assistant, defaultMaxBodyRunes))
	}
	if len(parts) == 0 {
		parts = append(parts, "本轮处理已结束。")
	}

	title := "Codex 已完成一轮处理"
	if project != "" {
		title = fmt.Sprintf("%s [%s]", title, project)
	}

	return notify.Notification{
		AppID:   defaultAppID,
		Title:   title,
		Message: strings.Join(parts, "\n"),
		Group:   fallback(env.ThreadID, "default"),
		Tag:     fallback(env.TurnID, hashTag(env.RawJSON)),
	}
}

func renderPermissionRequest(env event.Envelope) notify.Notification {
	project := projectName(env.Cwd)
	toolName := event.GetString(env.Raw, "tool_name", "toolName")
	permissionMode := event.GetString(env.Raw, "permission_mode", "permissionMode")
	model := event.GetString(env.Raw, "model")
	description := getNestedString(env.Raw, "tool_input", "description")
	command := getNestedString(env.Raw, "tool_input", "command")

	parts := make([]string, 0, 4)
	if toolName != "" {
		parts = append(parts, "Tool: "+toolName)
	}
	if command != "" {
		parts = append(parts, "Command: "+cleanText(command, defaultMaxBodyRunes))
	}
	if description != "" {
		parts = append(parts, "Reason: "+cleanText(description, 160))
	}
	if permissionMode != "" {
		parts = append(parts, "Mode: "+permissionMode)
	}
	if model != "" {
		parts = append(parts, "Model: "+model)
	}
	if len(parts) == 0 {
		parts = append(parts, "Codex is requesting permission.")
	}

	title := "Codex Permission Request"
	if project != "" {
		title = fmt.Sprintf("%s [%s]", title, project)
	}

	return notify.Notification{
		AppID:   defaultAppID,
		Title:   title,
		Message: strings.Join(parts, "\n"),
		Group:   fallback(env.ThreadID, "default"),
		Tag:     fallback(env.TurnID, hashTag(env.RawJSON)),
	}
}

func renderDefault(env event.Envelope) notify.Notification {
	eventType := fallback(env.Type, "unknown")
	message := firstNonEmpty(
		event.GetString(env.Raw, "message"),
		event.GetString(env.Raw, "summary"),
		event.GetString(env.Raw, "text"),
		event.GetString(env.Raw, "status"),
	)

	if message == "" {
		message = summarizeRaw(env.Raw)
	}

	return notify.Notification{
		AppID:   defaultAppID,
		Title:   fmt.Sprintf("Codex 事件: %s", eventType),
		Message: cleanText(message, defaultMaxBodyRunes),
		Group:   fallback(env.ThreadID, "default"),
		Tag:     fallback(env.TurnID, hashTag(env.RawJSON)),
	}
}

func summarizeRaw(raw map[string]any) string {
	keys := []string{"type", "event_type", "eventType", "hook_event_name", "hookEventName", "client", "cwd", "thread-id", "thread_id", "threadId", "session_id", "sessionId", "turn-id", "turn_id", "turnId"}
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		if value := event.GetString(raw, key); value != "" {
			parts = append(parts, fmt.Sprintf("%s=%s", key, value))
		}
	}
	if len(parts) > 0 {
		return strings.Join(parts, ", ")
	}

	data, err := json.Marshal(raw)
	if err != nil {
		return "收到未知事件。"
	}
	return string(data)
}

func cleanText(input string, maxRunes int) string {
	text := strings.ReplaceAll(input, "`", "")
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cleaned = append(cleaned, line)
	}
	text = strings.Join(cleaned, " ")
	text = strings.Join(strings.Fields(text), " ")

	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	return string(runes[:maxRunes]) + "..."
}

func getNestedString(raw map[string]any, objectKey string, valueKeys ...string) string {
	value, ok := raw[objectKey]
	if !ok {
		return ""
	}
	nested, ok := value.(map[string]any)
	if !ok {
		return ""
	}
	return event.GetString(nested, valueKeys...)
}

func projectName(cwd string) string {
	if cwd == "" {
		return ""
	}
	base := filepath.Base(cwd)
	if base == "." || base == string(filepath.Separator) {
		return ""
	}
	return base
}

func hashTag(raw []byte) string {
	sum := sha1.Sum(raw)
	return hex.EncodeToString(sum[:8])
}

func fallback(value string, defaultValue string) string {
	if strings.TrimSpace(value) != "" {
		return value
	}
	return defaultValue
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
