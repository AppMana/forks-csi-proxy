package utils

import (
	"context"
	"os"
)

const MaxPathLengthWindows = 260

func RunPowershellCmd(command string, envs ...string) ([]byte, error) {
	return RunPowershellCmdContext(context.Background(), command, envs...)
}

func powershellEnvironment(envs ...string) []string {
	return append(os.Environ(), envs...)
}
