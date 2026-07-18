//go:build windows
// +build windows

package utils

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRunPowershellCmdContextKillsDescendants(t *testing.T) {
	marker := filepath.Join(t.TempDir(), "orphaned.txt")
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := RunPowershellCmdContext(ctx,
		`$child = Start-Process powershell -PassThru -ArgumentList '-NoProfile','-Command','Start-Sleep -Seconds 3; Set-Content -Path $env:JOB_TEST_MARKER -Value orphaned'; Wait-Process -Id $child.Id`,
		"JOB_TEST_MARKER="+marker,
	)
	require.True(t, errors.Is(err, context.DeadlineExceeded), err)

	// The child would create the marker after three seconds if closing the Job
	// Object did not terminate the complete PowerShell process tree.
	time.Sleep(4 * time.Second)
	_, err = os.Stat(marker)
	require.True(t, os.IsNotExist(err), "cancelled PowerShell descendant survived and wrote %s", marker)
}
