//go:build wails

package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"sensitive-filter/pkg/gui"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	svc := gui.NewService(gui.DefaultWordsPath)
	app := NewApp(svc)

	err := wails.Run(&options.App{
		Title:       "敏感词过滤系统",
		Width:       1480,
		Height:      920,
		MinWidth:    1280,
		MinHeight:   760,
		AssetServer: &assetserver.Options{Assets: assets},
		OnStartup:   app.startup,
		Bind: []interface{}{
			app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}
