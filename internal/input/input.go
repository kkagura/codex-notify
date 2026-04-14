package input

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

var ErrEmptyPayload = errors.New("empty payload: expected JSON from first positional arg, --payload, or stdin")

func ReadPayload(payloadFlag string, positionalArgs []string) ([]byte, error) {
	if strings.TrimSpace(payloadFlag) != "" {
		return []byte(payloadFlag), nil
	}

	if len(positionalArgs) > 0 && strings.TrimSpace(positionalArgs[0]) != "" {
		return []byte(positionalArgs[0]), nil
	}

	info, err := os.Stdin.Stat()
	if err != nil {
		return nil, fmt.Errorf("inspect stdin: %w", err)
	}
	if info.Mode()&os.ModeCharDevice == 0 {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("read stdin: %w", err)
		}
		if strings.TrimSpace(string(data)) != "" {
			return data, nil
		}
	}

	return nil, ErrEmptyPayload
}
