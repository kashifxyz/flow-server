package jobs

import (
	"context"

	"github.com/rs/zerolog"
)

type Worker struct {
	Log zerolog.Logger
}

func (w Worker) Run(ctx context.Context) {
	w.Log.Info().Msg("job worker started")
	<-ctx.Done()
	w.Log.Info().Msg("job worker stopped")
}

type Scheduler struct {
	Log    zerolog.Logger
	Worker Worker
}

func (s Scheduler) Run(ctx context.Context) {
	s.Log.Info().Msg("scheduler started")
	s.Worker.Run(ctx)
	s.Log.Info().Msg("scheduler stopped")
}
