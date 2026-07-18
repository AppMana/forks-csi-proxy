package iscsi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConnectTargetIsIdempotent(t *testing.T) {
	original := runPowershellCmd
	t.Cleanup(func() { runPowershellCmd = original })

	var command string
	var environment []string
	runPowershellCmd = func(cmd string, envs ...string) ([]byte, error) {
		command = cmd
		environment = append([]string(nil), envs...)
		return nil, nil
	}

	err := (APIImplementor{}).ConnectTarget(
		&TargetPortal{Address: "192.0.2.10", Port: 3261},
		"iqn.2026-07.io.longhorn:test", "NONE", "", "",
	)
	require.NoError(t, err)
	require.Contains(t, command, "Get-IscsiTarget -NodeAddress ${Env:iscsi_target_iqn}")
	require.Contains(t, command, "Where-Object { $_.IsConnected }")
	require.Contains(t, command, "if ($connected.Count -eq 0) { Connect-IscsiTarget")
	require.True(t, strings.HasSuffix(command, " | Out-Null }"), command)
	require.ElementsMatch(t, []string{
		"iscsi_tp_address=192.0.2.10",
		"iscsi_tp_port=3261",
		"iscsi_target_iqn=iqn.2026-07.io.longhorn:test",
		"iscsi_auth_type=NONE",
		"iscsi_chap_user=",
		"iscsi_chap_secret=",
	}, environment)
}

func TestConnectTargetKeepsChapInsideConditional(t *testing.T) {
	original := runPowershellCmd
	t.Cleanup(func() { runPowershellCmd = original })

	var command string
	runPowershellCmd = func(cmd string, _ ...string) ([]byte, error) {
		command = cmd
		return nil, nil
	}

	err := (APIImplementor{}).ConnectTarget(
		&TargetPortal{Address: "192.0.2.10", Port: 3260},
		"iqn.2026-07.io.longhorn:chap", "ONEWAYCHAP", "user", "secret",
	)
	require.NoError(t, err)
	require.Contains(t, command, "-ChapUsername ${Env:iscsi_chap_user}")
	require.Contains(t, command, "-ChapSecret ${Env:iscsi_chap_secret} | Out-Null }")
}
