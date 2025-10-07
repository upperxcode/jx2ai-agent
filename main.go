package api

import (
	"embed"
	"log"

	"github.com/johnxcode/jx2ai-agent/api"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/logger"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Create an instance of the app structure
	app, err := api.NewApp()
	if err != nil {
		log.Fatalf("Erro ao inicializar a aplicação: %v", err)
	}

	// Configuração da aplicação Wails
	appOptions := &options.App{
		Title:     "jxai-agent",
		Width:     1024,
		Height:    768,
		MinWidth:  600,
		MinHeight: 400,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		DisableResize:     false,
		Fullscreen:        false,
		Frameless:         false,
		StartHidden:       false,
		HideWindowOnClose: false,

		LogLevel:         logger.DEBUG,
		OnStartup:        app.Startup,
		BackgroundColour: &options.RGBA{R: 0, G: 0, B: 0, A: 0}, // Transparente

		Menu:   nil,
		Logger: nil,
		//OnStartup:         app.Startup,
		//OnDomReady:        app.DomReady,
		//OnBeforeClose:     app.BeforeClose,
		//OnShutdown:        app.Shutdown,
		WindowStartState: options.Normal,
		Bind: []any{
			app,
		},
		// Windows platform specific options
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
			// DisableFramelessWindowDecorations: false,
			WebviewUserDataPath: "",
		},
		// Mac platform specific options
		Mac: &mac.Options{
			TitleBar: &mac.TitleBar{
				TitlebarAppearsTransparent: true,
				HideTitle:                  false,
				HideTitleBar:               false,
				FullSizeContent:            false,
				UseToolbar:                 false,
				HideToolbarSeparator:       true,
			},
			Appearance:           mac.NSAppearanceNameDarkAqua,
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			About: &mac.AboutInfo{
				Title:   "jxai-agent",
				Message: "",
			},
		},
	}

	err = wails.Run(appOptions)

	if err != nil {
		log.Fatal(err)
	}
}
