package scheduler_k3s

import (
	"context"
	"fmt"

	"github.com/dokku/dokku/plugins/common"
)

func copyRegistryToNode(ctx context.Context, remoteHost string) error {
	common.LogInfo1Quiet(fmt.Sprintf("Updating k3s registry configuration on %s", remoteHost))
	sftpCmd, err := common.CallSftpCopy(common.SftpCopyInput{
		AllowUknownHosts: true,
		DestinationPath:  "/tmp/registries.yaml",
		RemoteHost:       remoteHost,
		SourcePath:       RegistryConfigPath,
	})
	if err != nil {
		return fmt.Errorf("Error copying registries.yaml to remote host: %w", err)
	}

	if sftpCmd.ExitErr != nil {
		return fmt.Errorf("Error copying registries.yaml to remote host: %s", sftpCmd.ExitErr.Error())
	}

	common.LogVerboseQuiet("Moving k3s registry configuration into place")
	mvCmd, err := common.CallSshCommand(common.SshCommandInput{
		AllowUknownHosts: true,
		Args:             []string{"/tmp/registries.yaml", RegistryConfigPath},
		Command:          "mv",
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call mv command over ssh: %w", err)
	}

	if mvCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from mv command over ssh: %d", mvCmd.ExitCode)
	}

	common.LogVerboseQuiet("Updating k3s registry permissions")
	chmodCmd, err := common.CallSshCommand(common.SshCommandInput{
		AllowUknownHosts: true,
		Args:             []string{"0644", RegistryConfigPath},
		Command:          "chmod",
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call chmod command over ssh: %w", err)
	}

	if chmodCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from chmod command over ssh: %d", chmodCmd.ExitCode)
	}

	common.LogVerboseQuiet("Updating k3s registry ower")
	chownCmd, err := common.CallSshCommand(common.SshCommandInput{
		AllowUknownHosts: true,
		Args:             []string{"root:root", RegistryConfigPath},
		Command:          "chown",
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call chown command over ssh: %w", err)
	}

	if chownCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from chown command over ssh: %d", chownCmd.ExitCode)
	}

	return nil
}
