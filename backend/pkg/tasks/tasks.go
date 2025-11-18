package tasks

import (
	"context"
	"sync"

	"github.com/rs/xid"
)

type Task interface {
	Handle(ctx context.Context)
}

type Handler struct {
	id     string
	cancel context.CancelFunc
	done   func()
}

func (h *Handler) Execute(ctx context.Context, task Task) {
	ctx, h.cancel = context.WithCancel(ctx)
	go func(ctx context.Context, task Task, done func()) {
		task.Handle(ctx)
		done()
	}(ctx, task, h.done)
}

func (h *Handler) Cancel() {
	h.cancel()
	h.done()
}

func New() *Manager {
	return &Manager{
		handlers: sync.Map{},
	}
}

type Manager struct {
	handlers sync.Map
}

func (manager *Manager) Execute(ctx context.Context, task Task) (pid string) {
	pid = xid.New().String()
	handler := &Handler{
		id: pid,
		done: func() {
			manager.handlers.Delete(pid)
		},
	}
	handler.Execute(ctx, task)
	manager.handlers.Store(pid, handler)
	return
}

func (manager *Manager) Cancel(pid string) (ok bool) {
	handler0, has := manager.handlers.Load(pid)
	if !has {
		return
	}
	handler := handler0.(*Handler)
	handler.cancel()
	ok = true
	return
}

func (manager *Manager) Shutdown() {
	manager.handlers.Range(func(k, v interface{}) bool {
		v.(*Handler).cancel()
		return true
	})
	manager.handlers.Clear()
}
