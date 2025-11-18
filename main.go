package main

import (
	"context"
	"embed"

	"DarkestDungeonModBoxLite/backend"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app := backend.New()

	// Create application with options
	err := wails.Run(&options.App{
		Title:     "DarkestDungeonModBoxLite",
		MinWidth:  1024,
		MinHeight: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup: app.Startup,
		OnBeforeClose: func(ctx context.Context) (prevent bool) {
			app.Shutdown(ctx)
			return false
		},
		OnShutdown: app.Shutdown,
		Bind:       app.Binds(),
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: app.Id(),
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				app.Show()
			},
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
