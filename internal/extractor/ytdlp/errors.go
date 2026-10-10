package ytdlp

import (
	"fmt"
	"strings"
)

type CommandError struct {
	Err                error
	Stderr             string
	UsesBrowserCookies bool
}

func (e *CommandError) Error() string {
	if e.UsesBrowserCookies {
		message := strings.ToLower(e.Stderr)
		switch {
		case strings.Contains(message, "not a bot"), strings.Contains(message, "too many requests"), strings.Contains(message, "429"):
			return "O YouTube ainda está bloqueando esta sessão. Acesse o site no navegador escolhido e aguarde antes de tentar novamente."
		case strings.Contains(message, "cookie"), strings.Contains(message, "dpapi"), strings.Contains(message, "decrypt"), strings.Contains(message, "keyring"):
			return "Não foi possível ler os cookies do navegador. Acesse o YouTube nesse navegador, feche-o e tente novamente. Se persistir, selecione Firefox nas configurações; a criptografia do navegador pode impedir a leitura."
		default:
			return "Não foi possível concluir a operação com o yt-dlp. Verifique se o vídeo está acessível no navegador escolhido e tente novamente."
		}
	}
	message := strings.TrimSpace(e.Stderr)

	if message == "" {
		return fmt.Sprintf(
			"yt-dlp command failed: %v",
			e.Err,
		)
	}

	return fmt.Sprintf(
		"yt-dlp command failed: %s",
		message,
	)
}

func (e *CommandError) Unwrap() error {
	return e.Err
}

// UserMessage provides safe guidance for the application's job failure view.
func (e *CommandError) UserMessage() string {
	if e.UsesBrowserCookies {
		return e.Error()
	}
	return "Não foi possível concluir a operação."
}
