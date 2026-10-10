//go:build !linux && !windows

package browser

import (
	"context"
	"fmt"
)

func systemBrowser(context.Context) (string, error) {
	return "", fmt.Errorf("selecione um navegador nas configurações nesta plataforma")
}
