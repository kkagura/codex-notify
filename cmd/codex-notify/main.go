package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	"codex-notify/internal/dedupe"
	"codex-notify/internal/event"
	"codex-notify/internal/input"
	"codex-notify/internal/logging"
	"codex-notify/internal/notify/windows"
	"codex-notify/internal/render"
)

const (
	exitOK          = 0
	exitNotifyError = 1
	exitJSONError   = 2
	exitInitError   = 3
)

func main() {
	os.Exit(run())
}

func run() int {
	var payload string
	var verbose bool
	var silent bool

	flag.StringVar(&payload, "payload", "", "JSON payload")
	flag.BoolVar(&verbose, "verbose", false, "enable verbose logging")
	flag.BoolVar(&silent, "silent", false, "send notification without sound")
	flag.Parse()

	app, err := logging.NewApp(verbose)
	if err != nil {
		fmt.Fprintf(os.Stderr, "init logging: %v\n", err)
		return exitInitError
	}

	raw, err := input.ReadPayload(payload, flag.Args())
	if err != nil {
		app.Logger.Error().Err(err).Msg("read input payload failed")
		fmt.Fprintln(os.Stderr, err.Error())
		if errors.Is(err, input.ErrEmptyPayload) {
			return exitJSONError
		}
		return exitJSONError
	}

	env, err := event.Parse(raw)
	if err != nil {
		app.Logger.Error().Err(err).Bytes("raw", raw).Msg("parse event payload failed")
		fmt.Fprintln(os.Stderr, err.Error())
		return exitJSONError
	}

	logEvent := app.Logger.With().
		Str("type", env.Type).
		Str("thread_id", env.ThreadID).
		Str("turn_id", env.TurnID).
		Str("client", env.Client).
		Str("cwd", env.Cwd).
		Logger()

	logEvent.Info().Msg("event parsed")

	key := dedupe.Key(env)
	store := dedupe.NewStore(app.Paths.StateFile)
	seen, err := store.Seen(key)
	if err != nil {
		logEvent.Error().Err(err).Str("dedupe_key", key).Msg("check dedupe state failed")
	} else if seen {
		logEvent.Info().Str("dedupe_key", key).Msg("event skipped due to dedupe")
		fmt.Printf("ok: duplicate skipped key=%s\n", key)
		return exitOK
	}

	notification := render.Build(env)
	if silent {
		notification.Silent = true
	}

	notifier := windows.NewNotifier()
	if err := notifier.Send(context.Background(), notification); err != nil {
		logEvent.Error().Err(err).Str("title", notification.Title).Msg("send notification failed")
		fmt.Fprintf(os.Stderr, "send notification failed: %v\n", err)
		return exitNotifyError
	}

	if err := store.Mark(key); err != nil {
		logEvent.Error().Err(err).Str("dedupe_key", key).Msg("persist dedupe state failed")
	} else {
		logEvent.Info().Str("dedupe_key", key).Msg("notification sent")
	}

	fmt.Printf("ok: notified turn-id=%s\n", env.TurnID)
	return exitOK
}
