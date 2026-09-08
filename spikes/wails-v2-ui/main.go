package main

import (
	"embed"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	bridge := NewBridge()
	if err := wails.Run(&options.App{
		Title:     "ECO UI Architecture Spike — NON-PRODUCTION",
		Width:     1180,
		Height:    760,
		MinWidth:  860,
		MinHeight: 560,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		OnStartup:  bridge.startup,
		OnShutdown: bridge.shutdown,
		Bind:       []interface{}{bridge},
	}); err != nil {
		println("ECO Wails spike error:", err.Error())
	}
}
