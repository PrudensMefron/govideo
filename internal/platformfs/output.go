package platformfs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func AvailablePath(dir, base, ext string) (string, error) {
	if strings.TrimSpace(dir) == "" {
		return "", fmt.Errorf("destination is empty")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("create destination: %w", err)
	}
	base = sanitize(base)
	ext = strings.TrimPrefix(ext, ".")
	for n := 0; n < 10000; n++ {
		suffix := ""
		if n > 0 {
			suffix = fmt.Sprintf(" (%d)", n)
		}
		p := filepath.Join(dir, base+suffix+"."+ext)
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return p, nil
		}
	}
	return "", fmt.Errorf("cannot allocate output name")
}
func sanitize(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "media"
	}
	return strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\\|?*`, r) || r < 32 {
			return '_'
		}
		return r
	}, s)
}
