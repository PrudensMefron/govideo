package deps

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PrudensMefron/govideo/internal/process"
)

type State string

const (
	Missing    State = "missing"
	Installing State = "installing"
	Ready      State = "ready"
	Invalid    State = "invalid"
	Failed     State = "failed"
)

type Status struct {
	Name    string `json:"name"`
	State   State  `json:"state"`
	Version string `json:"version,omitempty"`
	Path    string `json:"path,omitempty"`
	Error   string `json:"error,omitempty"`
	License string `json:"license,omitempty"`
}
type Asset struct {
	Name, URL, SHA256 string
	Mode              os.FileMode
}
type Manager struct {
	mu       sync.Mutex
	statuses map[string]Status
	client   *http.Client
	root     string
	notify   func(Status)
}

func DataRoot() (string, error) {
	if runtime.GOOS == "windows" {
		p := os.Getenv("LOCALAPPDATA")
		if p == "" {
			return "", fmt.Errorf("LOCALAPPDATA is not set")
		}
		return filepath.Join(p, "GoVideo"), nil
	}
	if p := os.Getenv("XDG_DATA_HOME"); p != "" {
		return filepath.Join(p, "govideo"), nil
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(h, ".local", "share", "govideo"), nil
}
func New(root string, notify func(Status)) *Manager {
	return &Manager{statuses: map[string]Status{}, client: &http.Client{Timeout: 10 * time.Minute}, root: root, notify: notify}
}
func YTDLPAsset() (Asset, error) {
	switch runtime.GOOS {
	case "linux":
		return Asset{"yt-dlp_linux", "https://github.com/yt-dlp/yt-dlp/releases/download/2026.08.19/yt-dlp_linux", "58162f9bfdc27458ea47bfcb311cf47028f17d8154a8bf7d689861d46399230a", 0755}, nil
	case "windows":
		return Asset{"yt-dlp.exe", "https://github.com/yt-dlp/yt-dlp/releases/download/2026.08.19/yt-dlp.exe", "66674953fe251b89f4d08c5f0e35e0728679bd67ab3d7d05c0562af101dd3e7a", 0755}, nil
	default:
		return Asset{}, fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
}
func (m *Manager) List() []Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]Status, 0, len(m.statuses))
	for _, s := range m.statuses {
		out = append(out, s)
	}
	return out
}
func (m *Manager) set(s Status) {
	m.statuses[s.Name] = s
	if m.notify != nil {
		m.notify(s)
	}
}
func (m *Manager) EnsureYTDLP(ctx context.Context) (Status, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	a, err := YTDLPAsset()
	if err != nil {
		return Status{}, err
	}
	path := filepath.Join(m.root, "bin", a.Name)
	if v, e := version(ctx, path); e == nil {
		s := Status{Name: "yt-dlp", State: Ready, Version: v, Path: path}
		m.set(s)
		return s, nil
	}
	m.set(Status{Name: "yt-dlp", State: Installing, Path: path})
	if err = m.install(ctx, a, path, 200<<20); err != nil {
		s := Status{Name: "yt-dlp", State: Failed, Path: path, Error: err.Error()}
		m.set(s)
		return s, err
	}
	v, err := version(ctx, path)
	if err != nil {
		s := Status{Name: "yt-dlp", State: Invalid, Path: path, Error: err.Error()}
		m.set(s)
		return s, err
	}
	s := Status{Name: "yt-dlp", State: Ready, Version: v, Path: path}
	m.set(s)
	return s, nil
}
func (m *Manager) DetectFFmpeg(ctx context.Context) Status {
	m.mu.Lock()
	defer m.mu.Unlock()
	p, err := exec.LookPath("ffmpeg")
	if err != nil {
		s := Status{Name: "ffmpeg", State: Missing, License: "FFmpeg LGPL/GPL — ffmpeg.org/legal.html"}
		m.set(s)
		return s
	}
	v, e := version(ctx, p)
	st := Ready
	if e != nil {
		st = Invalid
	}
	s := Status{Name: "ffmpeg", State: st, Version: v, Path: p, License: "FFmpeg — ffmpeg.org/legal.html"}
	if e != nil {
		s.Error = e.Error()
	}
	m.set(s)
	return s
}
func (m *Manager) install(ctx context.Context, a Asset, target string, max int64) error {
	if len(a.SHA256) != 64 {
		return fmt.Errorf("invalid checksum manifest")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		return err
	}
	res, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", a.Name, err)
	}
	defer res.Body.Close()
	if res.StatusCode/100 != 2 {
		return fmt.Errorf("download %s: HTTP %s", a.Name, res.Status)
	}
	if res.ContentLength > max {
		return fmt.Errorf("asset exceeds size limit")
	}
	if err = os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(target), ".install-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer os.Remove(tmp)
	h := sha256.New()
	n, copyErr := io.Copy(io.MultiWriter(f, h), io.LimitReader(res.Body, max+1))
	if copyErr == nil && n > max {
		copyErr = fmt.Errorf("asset exceeds size limit")
	}
	if copyErr == nil {
		copyErr = f.Sync()
	}
	if e := f.Close(); copyErr == nil {
		copyErr = e
	}
	if copyErr != nil {
		return copyErr
	}
	if !strings.EqualFold(hex.EncodeToString(h.Sum(nil)), a.SHA256) {
		return fmt.Errorf("checksum mismatch")
	}
	if err = os.Chmod(tmp, a.Mode); err != nil {
		return err
	}
	return os.Rename(tmp, target)
}
func version(ctx context.Context, path string) (string, error) {
	out, err := process.CommandContext(ctx, path, "-version").CombinedOutput()
	if err != nil {
		out, err = process.CommandContext(ctx, path, "--version").CombinedOutput()
	}
	if err != nil {
		return "", err
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	return line, nil
}
