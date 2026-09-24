//go:build !windows

package process

import (
	"context"
	"testing"
)

func TestCommandContextUsesDefaultAttributesOutsideWindows(t *testing.T) {
	cmd := CommandContext(context.Background(), "true")
	if cmd.SysProcAttr != nil {
		t.Fatalf("unexpected process attributes: %#v", cmd.SysProcAttr)
	}
}
