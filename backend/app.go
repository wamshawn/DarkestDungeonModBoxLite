package backend

import (
	"context"
	"sync/atomic"

	"DarkestDungeonModBoxLite/backend/services/box"

	"github.com/google/uuid"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

func New() *App {
	app := &App{}
	app.register(box.Load)
	return app
}

type App struct {
	running   atomic.Bool
	ctx       context.Context
	services  []any
	startups  []func(ctx context.Context)
	shutdowns []func(ctx context.Context)
}

func (app *App) register(fn func() (service any, startup func(ctx context.Context), shutdown func(ctx context.Context))) {
	service, startup, shutdown := fn()
	app.services = append(app.services, service)
	app.startups = append(app.startups, startup)
	app.shutdowns = append(app.shutdowns, shutdown)
}

func (app *App) Id() string {
	return uuid.New().String()
}

func (app *App) Startup(ctx context.Context) {
	app.ctx = ctx
	for _, startup := range app.startups {
		startup(ctx)
	}
	app.running.Store(true)
}

func (app *App) Shutdown(ctx context.Context) {
	if app.running.CompareAndSwap(true, false) {
		for _, shutdown := range app.shutdowns {
			shutdown(ctx)
		}
	}
}

func (app *App) Show() {
	runtime.WindowShow(app.ctx)
}

func (app *App) Binds() []any {
	return app.services
}
