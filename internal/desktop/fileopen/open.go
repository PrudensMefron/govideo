// Package fileopen launches the system media handler without a shell or console window.
package fileopen

import (
	"context"
	"time"

	"github.com/PrudensMefron/govideo/internal/process"
)

func Open(path string) error {
	binary, args := command(path)
	cmd := process.CommandContext(context.Background(), binary, args...)
	if err := cmd.Start(); err != nil {
		return err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	// Report immediate association/launch errors without waiting for the player to close.
	timer := time.NewTimer(500 * time.Millisecond)
	defer timer.Stop()
	select {
	case err := <-done:
		return err
	case <-timer.C:
		return nil
	}
}
