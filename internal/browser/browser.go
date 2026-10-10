// Package browser resolves a browser preference to a yt-dlp cookie source.
package browser

import (
	"context"
	"fmt"
	"strings"
)

func Validate(preference string) error {
	switch preference {
	case "", "system", "firefox", "chrome", "chromium", "edge", "brave", "opera", "vivaldi", "whale":
		return nil
	default:
		return fmt.Errorf("navegador inválido; selecione um navegador nas configurações")
	}
}

func Resolve(ctx context.Context, preference string) (string, error) {
	if err := Validate(preference); err != nil {
		return "", err
	}
	if preference != "" && preference != "system" {
		return preference, nil
	}
	return systemBrowser(ctx)
}

func fromIdentifier(identifier string) (string, error) {
	id := strings.ToLower(strings.TrimSpace(identifier))
	for _, item := range []struct{ marker, browser string }{
		{"firefox", "firefox"}, {"brave", "brave"}, {"vivaldi", "vivaldi"},
		{"msedge", "edge"}, {"microsoft-edge", "edge"}, {"msehtm", "edge"},
		{"chromium", "chromium"}, {"chrome", "chrome"},
		{"opera", "opera"}, {"whale", "whale"},
	} {
		if strings.Contains(id, item.marker) {
			return item.browser, nil
		}
	}
	return "", fmt.Errorf("o navegador padrão não é compatível com a leitura de cookies; selecione um navegador nas configurações")
}
