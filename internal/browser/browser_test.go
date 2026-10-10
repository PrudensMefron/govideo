package browser

import (
	"context"
	"testing"
)

func TestBrowserIdentifiers(t *testing.T) {
	for _, tt := range []struct{ id, want string }{
		{"org.mozilla.firefox.desktop", "firefox"}, {"firefox-esr.desktop", "firefox"},
		{"com.google.Chrome.desktop", "chrome"}, {"chromium-browser.desktop", "chromium"},
		{"com.brave.Browser.desktop", "brave"}, {"microsoft-edge.desktop", "edge"},
		{"FirefoxURL-123", "firefox"}, {"ChromeHTML", "chrome"},
		{"MSEdgeHTM", "edge"}, {"BraveHTML", "brave"}, {"VivaldiHTM", "vivaldi"},
		{"OperaStable", "opera"}, {"WhaleHTM", "whale"},
	} {
		t.Run(tt.id, func(t *testing.T) {
			got, err := fromIdentifier(tt.id)
			if err != nil || got != tt.want {
				t.Fatalf("got %q, %v; want %q", got, err, tt.want)
			}
		})
	}
	if _, err := fromIdentifier("unknown.desktop"); err == nil {
		t.Fatal("unsupported system browser must require an explicit selection")
	}
}

func TestExplicitSelection(t *testing.T) {
	for _, selected := range []string{"firefox", "chrome", "chromium", "edge", "brave", "opera", "vivaldi", "whale"} {
		got, err := Resolve(context.Background(), selected)
		if err != nil || got != selected {
			t.Fatalf("explicit browser must override system: %q, %v", got, err)
		}
	}
	for _, invalid := range []string{"--help", "chrome:custom-profile", "unknown", "firefox;echo secret"} {
		if _, err := Resolve(context.Background(), invalid); err == nil {
			t.Fatalf("invalid preference accepted: %q", invalid)
		}
	}
}
