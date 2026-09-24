package main

import (
	"context"
	"encoding/json"
	"fmt"
	appsvc "github.com/PrudensMefron/govideo/internal/app"
	"github.com/PrudensMefron/govideo/internal/core"
	"github.com/PrudensMefron/govideo/internal/deps"
	appupdate "github.com/PrudensMefron/govideo/internal/desktop/update"
	"github.com/wailsapp/wails/v3/pkg/application"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type AnalyzeRequest struct {
	URL string `json:"url"`
}
type DownloadDTO struct {
	URL          string            `json:"url"`
	Media        core.Media        `json:"media"`
	Mode         core.OutputMode   `json:"mode"`
	MaxHeight    int               `json:"maxHeight"`
	AudioFormat  core.AudioFormat  `json:"audioFormat"`
	AudioQuality core.AudioQuality `json:"audioQuality"`
	Destination  string            `json:"destination"`
}
type ConversionDTO struct {
	Paths        []string          `json:"paths"`
	AudioFormat  core.AudioFormat  `json:"audioFormat"`
	AudioQuality core.AudioQuality `json:"audioQuality"`
	Destination  string            `json:"destination"`
}
type Settings struct {
	VideoDirectory      string `json:"videoDirectory"`
	MusicDirectory      string `json:"musicDirectory"`
	ConversionDirectory string `json:"conversionDirectory"`
	Theme               string `json:"theme"`
}
type DesktopService struct {
	app          *application.App
	jobs         *appsvc.Service
	deps         *deps.Manager
	updates      *appupdate.Manager
	settingsPath string
	mu           sync.Mutex
	settings     Settings
}

func NewDesktopService(a *application.App, j *appsvc.Service, d *deps.Manager, updates *appupdate.Manager, root string) *DesktopService {
	s := &DesktopService{app: a, jobs: j, deps: d, updates: updates, settingsPath: filepath.Join(root, "state", "settings.json")}
	s.loadSettings()
	return s
}
func (s *DesktopService) AnalyzeURL(r AnalyzeRequest) (core.Media, error) {
	return s.jobs.AnalyzeURL(context.Background(), r.URL)
}
func (s *DesktopService) StartDownload(r DownloadDTO) (core.Job, error) {
	return s.jobs.StartDownload(core.Source{URL: r.URL}, r.Media, r.Mode, core.VideoQuality{MaxHeight: r.MaxHeight}, core.AudioOptions{Format: r.AudioFormat, Quality: r.AudioQuality}, r.Destination)
}
func (s *DesktopService) StartLocalConversions(r ConversionDTO) ([]core.Job, error) {
	return s.jobs.StartConversions(r.Paths, core.AudioOptions{Format: r.AudioFormat, Quality: r.AudioQuality}, r.Destination)
}
func (s *DesktopService) ListJobs(offset, limit int) []core.Job {
	return s.jobs.ListJobs(offset, limit)
}
func (s *DesktopService) CancelJob(id string) (core.Job, error) { return s.jobs.Cancel(core.JobID(id)) }
func (s *DesktopService) RetryJob(id string) (core.Job, error)  { return s.jobs.Retry(core.JobID(id)) }
func (s *DesktopService) ListDependencies() []deps.Status       { return s.deps.List() }
func (s *DesktopService) RetryDependency(name string) (deps.Status, error) {
	if name != "yt-dlp" {
		return deps.Status{}, fmt.Errorf("unknown dependency %q", name)
	}
	return s.deps.EnsureYTDLP(context.Background())
}
func (s *DesktopService) GetUpdateStatus() appupdate.Status { return s.updates.Status() }
func (s *DesktopService) CheckForUpdates() (appupdate.Status, error) {
	return s.updates.CheckAndInstall(context.Background())
}
func (s *DesktopService) SelectInputVideos() ([]string, error) {
	return s.app.Dialog.OpenFile().SetTitle("Selecionar vídeos").AddFilter("Vídeos", "*.mp4;*.mkv;*.webm;*.mov;*.avi;*.m4v").PromptForMultipleSelection()
}
func (s *DesktopService) SelectDestination(kind string) (string, error) {
	return s.app.Dialog.OpenFile().CanChooseFiles(false).CanChooseDirectories(true).CanCreateDirectories(true).SetTitle("Selecionar pasta de destino").PromptForSingleSelection()
}
func (s *DesktopService) OpenArtifactLocation(path string) error {
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("empty artifact path")
	}
	return s.app.Env.OpenFileManager(path, true)
}
func (s *DesktopService) GetSettings() Settings { s.mu.Lock(); defer s.mu.Unlock(); return s.settings }
func (s *DesktopService) UpdateSettings(v Settings) (Settings, error) {
	for _, p := range []string{v.VideoDirectory, v.MusicDirectory, v.ConversionDirectory} {
		if strings.TrimSpace(p) == "" {
			return Settings{}, fmt.Errorf("destination cannot be empty")
		}
		if err := os.MkdirAll(p, 0755); err != nil {
			return Settings{}, err
		}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.settings = v
	b, _ := json.MarshalIndent(v, "", "  ")
	if err := os.MkdirAll(filepath.Dir(s.settingsPath), 0755); err != nil {
		return Settings{}, err
	}
	tmp := s.settingsPath + ".tmp"
	if err := os.WriteFile(tmp, b, 0600); err != nil {
		return Settings{}, err
	}
	if err := os.Rename(tmp, s.settingsPath); err != nil {
		return Settings{}, err
	}
	return v, nil
}
func (s *DesktopService) loadSettings() {
	base, _ := os.UserHomeDir()
	downloads := filepath.Join(base, "Downloads", "GoVideo")
	s.settings = Settings{VideoDirectory: filepath.Join(downloads, "Videos"), MusicDirectory: filepath.Join(downloads, "Musicas"), ConversionDirectory: filepath.Join(downloads, "Conversoes"), Theme: "system"}
	b, err := os.ReadFile(s.settingsPath)
	if err == nil {
		_ = json.Unmarshal(b, &s.settings)
	}
}
