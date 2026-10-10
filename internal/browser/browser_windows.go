//go:build windows

package browser

import (
	"context"
	"fmt"

	"golang.org/x/sys/windows/registry"
)

func systemBrowser(ctx context.Context) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\Shell\Associations\UrlAssociations\https\UserChoice`, registry.QUERY_VALUE)
	if err != nil {
		return "", fmt.Errorf("não foi possível identificar o navegador padrão; selecione um navegador nas configurações")
	}
	defer key.Close()
	identifier, _, err := key.GetStringValue("ProgId")
	if err != nil {
		return "", fmt.Errorf("não foi possível identificar o navegador padrão; selecione um navegador nas configurações")
	}
	return fromIdentifier(identifier)
}
