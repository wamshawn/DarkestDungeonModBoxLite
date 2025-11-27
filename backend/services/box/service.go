package box

import (
	"context"
)

func Load() (service any, startup func(ctx context.Context), shutdown func(ctx context.Context)) {
	s := &Box{
		ctx:    nil,
		cancel: nil,
		db:     nil,
	}
	service = s
	startup = s.startup
	shutdown = s.shutdown
	return
}
