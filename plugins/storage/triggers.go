package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/dokku/dokku/plugins/common"
)

// TriggerInstall sets up the storage plugin on installation
func TriggerInstall() error {
	storageDir := GetStorageDirectory()

	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return fmt.Errorf("Unable to create storage directory: %s", err.Error())
	}

	if err := common.SetPermissions(common.SetPermissionInput{
		Filename: storageDir,
		Mode:     0755,
	}); err != nil {
		return fmt.Errorf("Unable to set storage directory permissions: %s", err.Error())
	}

	distro := detectDistro()
	if distro == "" {
		return nil
	}

	pluginPath := common.MustGetEnv("PLUGIN_AVAILABLE_PATH")
	chownScript := filepath.Join(pluginPath, "storage", "bin", "chown-storage-dir")

	sudoersFile := "/etc/sudoers.d/dokku-storage"
	content := fmt.Sprintf("%%dokku ALL=(ALL) NOPASSWD:%s *\n", chownScript)
	content += "Defaults env_keep += \"DOKKU_LIB_ROOT\"\n"

	if err := os.WriteFile(sudoersFile, []byte(content), 0440); err != nil {
		return fmt.Errorf("Unable to write sudoers file: %s", err.Error())
	}

	return nil
}

// detectDistro returns the Linux distribution name
func detectDistro() string {
	if runtime.GOOS != "linux" {
		return ""
	}

	distro := os.Getenv("DOKKU_DISTRO")
	if distro != "" {
		return distro
	}

	if common.FileExists("/etc/debian_version") {
		return "debian"
	}
	if common.FileExists("/etc/arch-release") {
		return "arch"
	}

	return ""
}

// TriggerStorageList outputs storage mounts for an app
func TriggerStorageList(appName string, phase string, format string) error {
	mounts, err := GetBindMounts(appName, phase)
	if err != nil {
		return err
	}

	if format == "json" {
		entries := []StorageListEntry{}
		for _, mount := range mounts {
			entries = append(entries, ParseMountPath(mount))
		}

		output, err := json.Marshal(entries)
		if err != nil {
			return err
		}
		fmt.Println(string(output))
	} else {
		for _, mount := range mounts {
			fmt.Println(mount)
		}
	}

	return nil
}
