package windows

import (
	"context"

	"codex-notify/internal/notify"

	"github.com/go-toast/toast"
)

type Notifier struct{}

func NewNotifier() *Notifier {
	return &Notifier{}
}

func (n *Notifier) Send(_ context.Context, item notify.Notification) error {
	t := toast.Notification{
		AppID:   item.AppID,
		Title:   item.Title,
		Message: item.Message,
		Icon:    item.Icon,
	}
	if item.Silent {
		t.Audio = toast.Silent
	}

	return t.Push()
}
