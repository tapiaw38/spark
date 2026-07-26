package commands

import (
	"context"
	"os/exec"
)

type Cmd = exec.Cmd

func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func Command(name string, args ...string) *Cmd {
	return exec.Command(name, args...)
}

func CommandContext(ctx context.Context, name string, args ...string) *Cmd {
	return exec.CommandContext(ctx, name, args...)
}
