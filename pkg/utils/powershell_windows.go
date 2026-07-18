//go:build windows
// +build windows

package utils

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
	"k8s.io/klog/v2"
)

// RunPowershellCmdContext runs PowerShell in a per-command Windows Job Object.
// Closing or terminating the job kills the complete descendant process tree,
// including CIM helper processes that can otherwise outlive a cancelled CSI
// RPC and serialize later iSCSI operations.
func RunPowershellCmdContext(ctx context.Context, command string, envs ...string) ([]byte, error) {
	job, err := newKillOnCloseJob()
	if err != nil {
		return nil, err
	}
	defer windows.CloseHandle(job)

	cmd := exec.Command("powershell", "-Mta", "-NoProfile", "-Command", command)
	cmd.Env = powershellEnvironment(envs...)
	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	klog.V(8).Infof("Executing command in Windows Job Object: %q", cmd.String())
	if err := cmd.Start(); err != nil {
		return output.Bytes(), err
	}

	process, err := windows.OpenProcess(
		windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE,
		false,
		uint32(cmd.Process.Pid),
	)
	if err != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return output.Bytes(), fmt.Errorf("open PowerShell process for Job Object: %w", err)
	}
	assignErr := windows.AssignProcessToJobObject(job, process)
	_ = windows.CloseHandle(process)
	if assignErr != nil {
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return output.Bytes(), fmt.Errorf("assign PowerShell process to Job Object: %w", assignErr)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return output.Bytes(), err
	case <-ctx.Done():
		_ = windows.TerminateJobObject(job, 1)
		<-done
		return output.Bytes(), ctx.Err()
	}
}

func newKillOnCloseJob() (windows.Handle, error) {
	job, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("create Windows Job Object: %w", err)
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{}
	limits.BasicLimitInformation.LimitFlags = windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE
	if _, err := windows.SetInformationJobObject(
		job,
		windows.JobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&limits)),
		uint32(unsafe.Sizeof(limits)),
	); err != nil {
		_ = windows.CloseHandle(job)
		return windows.InvalidHandle, fmt.Errorf("configure Windows Job Object: %w", err)
	}
	return job, nil
}
