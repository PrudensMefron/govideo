package update

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v3/pkg/updater"
	"github.com/wailsapp/wails/v3/pkg/updater/providers/github"
	"golang.org/x/mod/semver"
)

var (
	ErrDisabled = errors.New("automatic updates are disabled in development builds")
	ErrBusy     = errors.New("an update operation is already running")
)

type Status struct {
	Enabled          bool          `json:"enabled"`
	State            updater.State `json:"state"`
	CurrentVersion   string        `json:"currentVersion"`
	AvailableVersion string        `json:"availableVersion,omitempty"`
	Repository       string        `json:"repository"`
	Error            string        `json:"error,omitempty"`
}

// Manager is the desktop boundary around Wails' updater. The actual release
// lookup, download, checksum verification, swap and relaunch remain owned by
// Wails; this type adds GoVideo's release policy and serialises UI requests.
type Manager struct {
	engine           *updater.Updater
	version          string
	repository       string
	enabled          bool
	busy             atomic.Bool
	availableVersion atomic.Value
	lastError        atomic.Value
}

func New(engine *updater.Updater, version, repository string) (*Manager, error) {
	if engine == nil {
		return nil, errors.New("update engine is required")
	}
	if strings.TrimSpace(repository) == "" {
		return nil, errors.New("update repository is required")
	}

	provider, err := github.New(github.Config{
		Repository:    repository,
		ChecksumAsset: "SHA256SUMS",
		HTTPClient:    &http.Client{Timeout: 10 * time.Minute},
	})
	if err != nil {
		return nil, fmt.Errorf("configure GitHub update provider: %w", err)
	}
	if err := engine.Init(updater.Config{
		CurrentVersion: version,
		Providers:      []updater.Provider{requireChecksum(provider)},
	}); err != nil {
		return nil, fmt.Errorf("configure updater: %w", err)
	}

	m := &Manager{
		engine:     engine,
		version:    version,
		repository: repository,
		enabled:    isReleaseVersion(version),
	}
	m.availableVersion.Store("")
	m.lastError.Store("")
	return m, nil
}

func isReleaseVersion(version string) bool {
	version = strings.TrimSpace(strings.TrimPrefix(version, "v"))
	return version != "" && version != "0.0.0-dev" && semver.IsValid("v"+version)
}

func (m *Manager) Status() Status {
	return Status{
		Enabled:          m.enabled,
		State:            m.engine.State(),
		CurrentVersion:   m.version,
		AvailableVersion: m.availableVersion.Load().(string),
		Repository:       m.repository,
		Error:            m.lastError.Load().(string),
	}
}

// Check performs a non-interactive check. It never downloads an artifact.
func (m *Manager) Check(ctx context.Context) (Status, error) {
	if err := m.begin(); err != nil {
		return m.Status(), err
	}
	defer m.busy.Store(false)

	release, err := m.engine.Check(ctx)
	if err != nil {
		m.lastError.Store(err.Error())
		return m.Status(), err
	}
	m.lastError.Store("")
	if release == nil {
		m.availableVersion.Store("")
	} else {
		m.availableVersion.Store(release.Version)
	}
	return m.Status(), nil
}

// CheckAndInstall opens Wails' native updater window and runs the complete
// download/verification flow. Restart remains an explicit action in that
// window.
func (m *Manager) CheckAndInstall(ctx context.Context) (Status, error) {
	if err := m.begin(); err != nil {
		return m.Status(), err
	}
	defer m.busy.Store(false)

	err := m.engine.CheckAndInstall(ctx)
	if err != nil {
		m.lastError.Store(err.Error())
		return m.Status(), err
	}
	m.lastError.Store("")
	return m.Status(), nil
}

func (m *Manager) begin() error {
	if !m.enabled {
		return ErrDisabled
	}
	if !m.busy.CompareAndSwap(false, true) {
		return ErrBusy
	}
	return nil
}
