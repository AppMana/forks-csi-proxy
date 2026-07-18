//go:build !windows
// +build !windows

package utils

import (
	"context"
	"os/exec"

	"k8s.io/klog/v2"
)

func RunPowershellCmdContext(ctx context.Context, command string, envs ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "powershell", "-Mta", "-NoProfile", "-Command", command)
	cmd.Env = powershellEnvironment(envs...)
	klog.V(8).Infof("Executing command: %q", cmd.String())
	return cmd.CombinedOutput()
}
