package process

import (
	"context"
	"os/exec"
)

// CommandContext creates an external command with the platform-specific
// process attributes required by GoVideo. On Windows this always suppresses
// console windows, including when GoVideo itself was started outside a shell.
func CommandContext(ctx context.Context, name string, args ...string) *exec.Cmd {
	cmd := exec.CommandContext(ctx, name, args...)
	configure(cmd)
	return cmd
}
