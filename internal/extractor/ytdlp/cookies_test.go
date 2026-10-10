package ytdlp

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCookieResolverAndSafeErrors(t *testing.T) {
	a := New("yt-dlp")
	selected := "firefox"
	a.SetCookieBrowser(func(context.Context) (string, error) { return selected, nil })
	for _, browser := range []string{"firefox", "chrome"} {
		selected = browser
		args, err := a.cookieArgs(context.Background())
		if err != nil || len(args) != 2 || args[0] != "--cookies-from-browser" || args[1] != browser {
			t.Fatalf("configuration not refreshed: %v %v", args, err)
		}
	}
	wantErr := errors.New("browser unavailable")
	a.SetCookieBrowser(func(context.Context) (string, error) { return "", wantErr })
	if _, err := a.cookieArgs(context.Background()); !errors.Is(err, wantErr) {
		t.Fatalf("resolver error was lost: %v", err)
	}
	for _, stderr := range []string{"cookie error session=SECRET", "decrypt DPAPI SECRET", "not a bot SECRET", "unexpected error SECRET"} {
		err := &CommandError{Err: wantErr, Stderr: stderr, UsesBrowserCookies: true}
		if strings.Contains(err.Error(), "SECRET") || !errors.Is(err, wantErr) {
			t.Fatalf("stderr leaked or underlying error lost: %v", err)
		}
	}
}
