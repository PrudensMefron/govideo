package main

import (
	"context"
	"embed"
	appsvc "github.com/PrudensMefron/govideo/internal/app"
	"github.com/PrudensMefron/govideo/internal/core"
	"github.com/PrudensMefron/govideo/internal/deps"
	"github.com/PrudensMefron/govideo/internal/extractor/ytdlp"
	"github.com/PrudensMefron/govideo/internal/media/ffmpeg"
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
	"log"
	"path/filepath"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	application.RegisterEvent[core.Job]("job:updated")
	application.RegisterEvent[deps.Status]("dependency:updated")
	application.RegisterEvent[[]string]("files:dropped")
}
func main() {
	a := application.New(application.Options{Name: "GoVideo", Description: "Baixe e converta mídia com segurança", Assets: application.AssetOptions{Handler: application.AssetFileServerFS(assets)}})
	root, err := deps.DataRoot()
	if err != nil {
		log.Fatal(err)
	}
	dm := deps.New(root, func(s deps.Status) { a.Event.Emit("dependency:updated", s) })
	ytAsset, _ := deps.YTDLPAsset()
	ytPath := filepath.Join(root, "bin", ytAsset.Name)
	ex := ytdlp.New(ytPath)
	jobs := appsvc.New(ex, ex, ffmpeg.New("", ""), filepath.Join(root, "state", "jobs.json"), 2, func(j core.Job) { a.Event.Emit("job:updated", j) })
	service := NewDesktopService(a, jobs, dm, root)
	a.RegisterService(application.NewService(service))
	win := a.Window.NewWithOptions(application.WebviewWindowOptions{Title: "GoVideo", Width: 1180, Height: 760, MinWidth: 760, MinHeight: 560, EnableFileDrop: true, BackgroundColour: application.NewRGB(244, 248, 255), URL: "/"})
	win.OnWindowEvent(events.Common.WindowFilesDropped, func(e *application.WindowEvent) { a.Event.Emit("files:dropped", e.Context().DroppedFiles()) })
	go func() { dm.DetectFFmpeg(context.Background()); _, _ = dm.EnsureYTDLP(context.Background()) }()
	if err = a.Run(); err != nil {
		log.Fatal(err)
	}
}
