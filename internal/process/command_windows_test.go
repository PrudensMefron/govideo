//go:build windows

package process

import (
	"context"
	"testing"
)

func TestCommandContextHidesWindowsConsole(t *testing.T) {
	cmd := CommandContext(context.Background(), "cmd.exe", "/c", "exit", "0")
	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr was not configured")
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("Windows console is not hidden")
	}
	if cmd.SysProcAttr.CreationFlags&createNoWindow == 0 {
		t.Fatal("CREATE_NO_WINDOW is not set")
	}
}
