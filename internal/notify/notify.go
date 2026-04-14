package notify

import "context"

type Notification struct {
	AppID   string
	Title   string
	Message string
	Icon    string
	Group   string
	Tag     string
	Silent  bool
}

type Notifier interface {
	Send(ctx context.Context, n Notification) error
}
