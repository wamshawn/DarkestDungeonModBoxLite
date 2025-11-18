package box

import (
	"context"

	"DarkestDungeonModBoxLite/backend/pkg/databases"
	"DarkestDungeonModBoxLite/backend/pkg/tasks"
)

func New() *Box {
	return &Box{
		ctx:     nil,
		cancel:  nil,
		manager: nil,
		db:      nil,
	}
}

type Box struct {
	ctx     context.Context
	cancel  context.CancelFunc
	manager *tasks.Manager
	db      *databases.Database
}

func (bx *Box) startup(ctx context.Context) {
	bx.ctx, bx.cancel = context.WithCancel(ctx)
	bx.manager = tasks.New()
	return
}

func (bx *Box) shutdown(_ context.Context) {
	bx.cancel()
	bx.closeDB()
	if bx.manager != nil {
		bx.manager.Shutdown()
	}
	return
}

func (bx *Box) database() (*databases.Database, error) {
	if bx.db == nil {
		return nil, ErrOpened
	}
	return bx.db, nil
}
