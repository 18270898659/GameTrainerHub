package main

import (
	"embed"
	"log"

	"GameTrainerHub/internal/app"
	"GameTrainerHub/internal/config"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	cfg := config.GetConfig()
	appInstance, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Fatal error creating application: %v", err)
	}

	err = wails.Run(&options.App{
		Title:     "Game Trainer Hub - PC 游戏修改器聚合管理客户端",
		Width:     1280,
		Height:    820,
		MinWidth:  1024,
		MinHeight: 680,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 20, B: 24, A: 255},
		OnStartup:        appInstance.Startup,
		OnShutdown:       appInstance.Shutdown,
		OnBeforeClose:    appInstance.OnBeforeClose,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: "GameTrainerHub-Instance-Lock",
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				if appInstance.GetContext() != nil {
					runtime.WindowShow(appInstance.GetContext())
					runtime.WindowUnminimise(appInstance.GetContext())
				}
			},
		},
		Bind: []any{
			appInstance,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			Theme:                windows.Dark,
		},
	})

	if err != nil {
		log.Fatalf("Wails application error: %v", err)
	}
}

