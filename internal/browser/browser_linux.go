//go:build linux

package browser

import (
	"context"
	"fmt"
	"time"

	"github.com/PrudensMefron/govideo/internal/process"
)

func systemBrowser(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	output, err := process.CommandContext(ctx, "xdg-mime", "query", "default", "x-scheme-handler/https").Output()
	if err != nil {
		return "", fmt.Errorf("não foi possível identificar o navegador padrão; selecione um navegador nas configurações")
	}
	return fromIdentifier(string(output))
}
