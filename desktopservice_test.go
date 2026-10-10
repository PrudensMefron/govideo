package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBrowserSettingsMigrationAndPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	if err := os.WriteFile(path, []byte(`{"theme":"dark"}`), 0600); err != nil {
		t.Fatal(err)
	}
	s := &DesktopService{settingsPath: path}
	s.loadSettings()
	settings := s.GetSettings()
	if settings.CookieBrowser != "system" || settings.Theme != "dark" {
		t.Fatalf("old settings did not retain system browser default: %+v", settings)
	}
	settings.VideoDirectory = filepath.Join(dir, "videos")
	settings.MusicDirectory = filepath.Join(dir, "music")
	settings.ConversionDirectory = filepath.Join(dir, "conversions")
	settings.CookieBrowser = "firefox"
	if _, err := s.UpdateSettings(settings); err != nil {
		t.Fatal(err)
	}
	reopened := &DesktopService{settingsPath: path}
	reopened.loadSettings()
	if got := reopened.GetSettings().CookieBrowser; got != "firefox" {
		t.Fatalf("browser selection did not persist: %q", got)
	}
	settings.CookieBrowser = "unknown"
	if _, err := s.UpdateSettings(settings); err == nil {
		t.Fatal("invalid browser accepted")
	}
	if got := s.GetSettings().CookieBrowser; got != "firefox" {
		t.Fatalf("invalid save changed current browser: %q", got)
	}
}
