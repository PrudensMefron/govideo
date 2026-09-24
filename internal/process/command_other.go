//go:build !windows

package process

import "os/exec"

func configure(_ *exec.Cmd) {}
