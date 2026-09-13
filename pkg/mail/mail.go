package mail

import (
	"context"

	"github.com/rs/zerolog"
)

type Message struct {
	To       string
	Template string
	Payload  map[string]any
}

type Sender interface {
	Send(ctx context.Context, msg Message) error
}

type LogSender struct {
	Log zerolog.Logger
}

func (s LogSender) Send(_ context.Context, msg Message) error {
	s.Log.Info().
		Str("to", msg.To).
		Str("template", msg.Template).
		Msg("mail queued")
	return nil
}
