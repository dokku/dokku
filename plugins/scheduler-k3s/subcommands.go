package scheduler_k3s

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"net/url"
	"os"
	"strings"

	"github.com/dokku/dokku/plugins/common"
	resty "github.com/go-resty/resty/v2"
)

// CommandInitialize initializes a k3s cluster on the local server
func CommandInitialize(taintScheduling bool) error {
	if err := isK3sInstalled(); err == nil {
		return fmt.Errorf("k3s already installed, cannot re-initialize k3s")
	}

	networkInterface := common.PropertyGetDefault("scheduler-k3s", "--global", "network-interface", "eth0")
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("Unable to get network interfaces: %w", err)
	}

	serverIp := ""
	for _, iface := range ifaces {
		if iface.Name == networkInterface {
			addr, err := iface.Addrs()
			if err != nil {
				return fmt.Errorf("Unable to get network addresses for interface %s: %w", networkInterface, err)
			}
			for _, a := range addr {
				if ipnet, ok := a.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						serverIp = ipnet.IP.String()
					}
				}
			}
		}
	}

	if len(serverIp) == 0 {
		return fmt.Errorf(fmt.Sprintf("Unable to determine server ip address from network-interface %s", networkInterface))
	}

	common.LogInfo1Quiet("Initializing k3s")

	common.LogInfo2Quiet("Updating apt")
	aptUpdateCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "apt-get",
		Args: []string{
			"update",
		},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get update command: %w", err)
	}
	if aptUpdateCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get update command: %d", aptUpdateCmd.ExitCode)
	}

	common.LogInfo2Quiet("Installing k3s dependencies")
	aptInstallCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "apt-get",
		Args: []string{
			"-y",
			"install",
			"ca-certificates",
			"curl",
			"open-iscsi",
			"nfs-common",
			"wireguard",
		},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get install command: %w", err)
	}
	if aptInstallCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get install command: %d", aptInstallCmd.ExitCode)
	}

	common.LogInfo2Quiet("Downloading k3s installer")
	client := resty.New()
	resp, err := client.R().
		Get("https://get.k3s.io")
	if err != nil {
		return fmt.Errorf("Unable to download k3s installer: %w", err)
	}
	if resp == nil {
		return fmt.Errorf("Missing response from k3s installer download: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("Invalid status code for k3s installer script: %d", resp.StatusCode())
	}

	f, err := os.CreateTemp("", "sample")
	if err != nil {
		return fmt.Errorf("Unable to create temporary file for k3s installer: %w", err)
	}
	defer os.Remove(f.Name())

	if err := f.Close(); err != nil {
		return fmt.Errorf("Unable to close k3s installer file: %w", err)
	}

	err = common.WriteSliceToFile(common.WriteSliceToFileInput{
		Filename: f.Name(),
		Lines:    strings.Split(resp.String(), "\n"),
		Mode:     os.FileMode(0755),
	})
	if err != nil {
		return fmt.Errorf("Unable to write k3s installer to file: %w", err)
	}

	fi, err := os.Stat(f.Name())
	if err != nil {
		return fmt.Errorf("Unable to get k3s installer file size: %w", err)
	}

	if fi.Size() == 0 {
		return fmt.Errorf("Invalid k3s installer filesize")
	}

	token := common.PropertyGet("scheduler-k3s", "--global", "token")
	if len(token) == 0 {
		n := 5
		b := make([]byte, n)
		if _, err := rand.Read(b); err != nil {
			return fmt.Errorf("Unable to generate random node name: %w", err)
		}

		token := strings.ToLower(fmt.Sprintf("%X", b))
		if err := CommandSet("--global", "token", token); err != nil {
			return fmt.Errorf("Unable to set k3s token: %w", err)
		}
	}

	nodeName := serverIp
	n := 5
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("Unable to generate random node name: %w", err)
	}
	nodeName = strings.ReplaceAll(strings.ToLower(fmt.Sprintf("ip-%s-%s", nodeName, fmt.Sprintf("%X", b))), ".", "-")

	args := []string{
		// initialize the cluster
		"--cluster-init",
		// disable local-storage
		"--disable", "local-storage",
		// expose etcd metrics
		"--etcd-expose-metrics",
		// use wireguard for flannel
		"--flannel-backend=wireguard-native",
		// bind controller-manager to all interfaces
		"--kube-controller-manager-arg", "bind-address=0.0.0.0",
		// bind proxy metrics to all interfaces
		"--kube-proxy-arg", "metrics-bind-address=0.0.0.0",
		// bind scheduler to all interfaces
		"--kube-scheduler-arg", "bind-address=0.0.0.0",
		// gc terminated pods
		"--kube-controller-manager-arg", "terminated-pod-gc-threshold=10",
		// specify the node name
		"--node-name", nodeName,
		// allow access for the dokku user
		"--write-kubeconfig-mode", "0644",
		// specify a token
		"--token", token,
	}
	if taintScheduling {
		args = append(args, "--node-taint", "node-role.kubernetes.io/master=true:NoSchedule")
	}

	common.LogInfo2Quiet("Running k3s installer")
	installerCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     f.Name(),
		Args:        args,
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s installer command: %w", err)
	}
	if installerCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s installer command: %d", installerCmd.ExitCode)
	}

	if err := common.TouchFile(RegistryConfigPath); err != nil {
		return fmt.Errorf("Error creating initial registries.yaml file")
	}

	common.LogInfo2Quiet("Setting registries.yaml permissions")
	registryAclCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "setfacl",
		Args: []string{
			"-m",
			"user:dokku:rwx",
			RegistryConfigPath,
		},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call setfacl command: %w", err)
	}
	if registryAclCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from setfacl command: %d", registryAclCmd.ExitCode)
	}

	common.LogInfo2Quiet("Installing longhorn")
	longhornCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "kubectl",
		Args: []string{
			"apply",
			"-f",
			"https://raw.githubusercontent.com/longhorn/longhorn/v1.4.0/deploy/longhorn.yaml",
		},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call kubectl command: %w", err)
	}
	if longhornCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from kubectl command: %d", longhornCmd.ExitCode)
	}

	common.LogInfo2Quiet("Installing k3s automatic upgrader")
	upgradeCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "kubectl",
		Args: []string{
			"apply",
			"-f",
			"https://github.com/rancher/system-upgrade-controller/releases/latest/download/system-upgrade-controller.yaml",
		},
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call kubectl command: %w", err)
	}
	if upgradeCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from kubectl command: %d", upgradeCmd.ExitCode)
	}

	common.LogVerboseQuiet("Done")

	return nil
}

// CommandClusterAdd adds a server to the k3s cluster
func CommandClusterAdd(role string, remoteHost string, allowUknownHosts bool, taintScheduling bool) error {
	if err := isK3sInstalled(); err != nil {
		return fmt.Errorf("k3s not installed, cannot join cluster")
	}

	if role != "server" && role != "worker" {
		return fmt.Errorf("Invalid server-type: %s", role)
	}

	token := common.PropertyGet("scheduler-k3s", "--global", "token")
	if len(token) == 0 {
		return fmt.Errorf("Missing k3s token")
	}

	if taintScheduling && role == "worker" {
		return fmt.Errorf("Taint scheduling can only be used on the server role")
	}

	networkInterface := common.PropertyGetDefault("scheduler-k3s", "--global", "network-interface", "eth0")
	ifaces, err := net.Interfaces()
	if err != nil {
		return fmt.Errorf("Unable to get network interfaces: %w", err)
	}

	serverIp := ""
	for _, iface := range ifaces {
		if iface.Name == networkInterface {
			addr, err := iface.Addrs()
			if err != nil {
				return fmt.Errorf("Unable to get network addresses for interface %s: %w", networkInterface, err)
			}
			for _, a := range addr {
				if ipnet, ok := a.(*net.IPNet); ok {
					if ipnet.IP.To4() != nil {
						serverIp = ipnet.IP.String()
					}
				}
			}
		}
	}

	if len(serverIp) == 0 {
		return fmt.Errorf(fmt.Sprintf("Unable to determine server ip address from network-interface %s", networkInterface))
	}

	// todo: check if k3s is installed on the remote host

	k3sVersionCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command: "k3s",
		Args: []string{
			"--version",
		},
		CaptureOutput: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s version command: %w", err)
	}
	if k3sVersionCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s --version command: %d", k3sVersionCmd.ExitCode)
	}

	k3sVersion := ""
	k3sVersionLines := strings.Split(string(k3sVersionCmd.Stdout), "\n")
	if len(k3sVersionLines) > 0 {
		k3sVersionParts := strings.Split(k3sVersionLines[0], " ")
		if len(k3sVersionParts) != 4 {
			return fmt.Errorf("Unable to get k3s version from k3s --version: %s", k3sVersionCmd.Stdout)
		}
		k3sVersion = k3sVersionParts[2]
	}
	common.LogDebug(fmt.Sprintf("k3s version: %s", k3sVersion))

	common.LogInfo1(fmt.Sprintf("Joining %s to k3s cluster as %s", remoteHost, role))
	common.LogInfo2Quiet("Updating apt")
	aptUpdateCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "apt-get",
		Args: []string{
			"update",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get update command over ssh: %w", err)
	}
	if aptUpdateCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get update command over ssh: %d", aptUpdateCmd.ExitCode)
	}

	common.LogInfo2Quiet("Installing k3s dependencies")
	aptInstallCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "apt-get",
		Args: []string{
			"-y",
			"install",
			"ca-certificates",
			"curl",
			"open-iscsi",
			"nfs-common",
			"wireguard",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call apt-get install command over ssh: %w", err)
	}
	if aptInstallCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from apt-get install command over ssh: %d", aptInstallCmd.ExitCode)
	}

	common.LogInfo2Quiet("Downloading k3s installer")
	curlTask, err := common.CallSshCommand(common.SshCommandInput{
		Command: "curl",
		Args: []string{
			"-o /tmp/k3s-installer.sh",
			"https://get.k3s.io",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call curl command over ssh: %w", err)
	}
	if curlTask.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from curl command over ssh: %d", curlTask.ExitCode)
	}

	common.LogInfo2Quiet("Setting k3s installer permissions")
	chmodCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command: "chmod",
		Args: []string{
			"0755",
			"/tmp/k3s-installer.sh",
		},
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call chmod command over ssh: %w", err)
	}
	if chmodCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from chmod command over ssh: %d", chmodCmd.ExitCode)
	}

	u, err := url.Parse(remoteHost)
	if err != nil {
		return fmt.Errorf("failed to parse remote host: %w", err)
	}

	nodeName := u.Hostname()
	n := 5
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Errorf("Unable to generate random node name: %w", err)
	}
	nodeName = strings.ReplaceAll(strings.ToLower(fmt.Sprintf("ip-%s-%s", nodeName, fmt.Sprintf("%X", b))), ".", "-")

	args := []string{
		// disable local-storage
		"--disable", "local-storage",
		// use wireguard for flannel
		"--flannel-backend=wireguard-native",
		// specify the node name
		"--node-name", nodeName,
		// server to connect to as the main
		"--server",
		fmt.Sprintf("https://%s:6443", serverIp),
		// specify a token
		"--token",
		token,
	}

	if role == "server" {
		args = append([]string{"server"}, args...)
		// expose etcd metrics
		args = append(args, "--etcd-expose-metrics")
		// bind controller-manager to all interfaces
		args = append(args, "--kube-controller-manager-arg", "bind-address=0.0.0.0")
		// bind proxy metrics to all interfaces
		args = append(args, "--kube-proxy-arg", "metrics-bind-address=0.0.0.0")
		// bind scheduler to all interfaces
		args = append(args, "--kube-scheduler-arg", "bind-address=0.0.0.0")
		// gc terminated pods
		args = append(args, "--kube-controller-manager-arg", "terminated-pod-gc-threshold=10")
		// allow access for the dokku user
		args = append(args, "--write-kubeconfig-mode", "0644")
	} else {
		// disable etcd on workers
		args = append(args, "--disable-etcd")
		// disable apiserver on workers
		args = append(args, "--disable-apiserver")
		// disable controller-manager on workers
		args = append(args, "--disable-controller-manager")
		// disable scheduler on workers
		args = append(args, "--disable-scheduler")
		// bind proxy metrics to all interfaces
		args = append(args, "--kube-proxy-arg", "metrics-bind-address=0.0.0.0")
	}
	if taintScheduling {
		args = append(args, "--node-taint", "node-role.kubernetes.io/master=true:NoSchedule")
	}

	common.LogInfo2Quiet(fmt.Sprintf("Adding %s k3s cluster", nodeName))
	joinCmd, err := common.CallSshCommand(common.SshCommandInput{
		Command:          "/tmp/k3s-installer.sh",
		Args:             args,
		AllowUknownHosts: allowUknownHosts,
		RemoteHost:       remoteHost,
		StreamStdio:      true,
		Sudo:             true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s installer command over ssh: %w", err)
	}
	if joinCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s installer command over ssh: %d", joinCmd.ExitCode)
	}
	ctx := context.Background()
	clientset, err := NewKubernetesClient()
	if err != nil {
		return fmt.Errorf("Unable to create kubernetes client: %w", err)
	}

	common.LogInfo2Quiet("Waiting for node to exist")
	nodes, err := waitForNodeToExist(ctx, WaitForNodeToExistInput{
		Clientset:  clientset,
		NodeName:   nodeName,
		RetryCount: 20,
	})
	if err != nil {
		return fmt.Errorf("Error waiting for pod to exist: %w", err)
	}
	if len(nodes) == 0 {
		return fmt.Errorf("Unable to find node after joining cluster, node will not be labeled appropriately and cannot access registry secrets")
	}

	err = copyRegistryToNode(ctx, remoteHost)
	if err != nil {
		return fmt.Errorf("Error copying registries.yaml to remote host: %w", err)
	}

	if role == "worker" {
		common.LogInfo2Quiet("Labeling node kubernetes.io/role=worker")
		err = clientset.LabelNode(ctx, LabelNodeInput{
			Name:  nodes[0].Name,
			Key:   "kubernetes.io/role",
			Value: "worker",
		})
		if err != nil {
			return fmt.Errorf("Unable to patch node: %w", err)
		}
	}

	common.LogInfo2Quiet("Annotating node with connection information")
	err = clientset.AnnotateNode(ctx, AnnotateNodeInput{
		Name:  nodes[0].Name,
		Key:   "dokku.com/remote-host",
		Value: remoteHost,
	})
	if err != nil {
		return fmt.Errorf("Unable to patch node: %w", err)
	}

	common.LogVerboseQuiet("Done")
	return nil
}

// CommandReport displays a scheduler-k3s report for one or more apps
func CommandReport(appName string, format string, infoFlag string) error {
	if len(appName) == 0 {
		apps, err := common.DokkuApps()
		if err != nil {
			return err
		}
		for _, appName := range apps {
			if err := ReportSingleApp(appName, format, infoFlag); err != nil {
				return err
			}
		}
		return nil
	}

	return ReportSingleApp(appName, format, infoFlag)
}

// CommandSet set or clear a scheduler-k3s property for an app
func CommandSet(appName string, property string, value string) error {
	common.CommandPropertySet("scheduler-k3s", appName, property, value, DefaultProperties, GlobalProperties)
	return nil
}

// CommandShowKubeconfig displays the kubeconfig file contents
func CommandShowKubeconfig() error {
	if !common.FileExists(KubeConfigPath) {
		return fmt.Errorf("Kubeconfig file does not exist: %s", KubeConfigPath)
	}

	b, err := os.ReadFile(KubeConfigPath)
	if err != nil {
		return fmt.Errorf("Unable to read kubeconfig file: %w", err)
	}

	fmt.Println(string(b))

	return nil
}

func CommandUninstall() error {
	if err := isK3sInstalled(); err != nil {
		return fmt.Errorf("k3s not installed, cannot uninstall")
	}

	common.LogInfo1("Uninstalling k3s")
	uninstallerCmd, err := common.CallExecCommand(common.ExecCommandInput{
		Command:     "/usr/local/bin/k3s-uninstall.sh",
		StreamStdio: true,
	})
	if err != nil {
		return fmt.Errorf("Unable to call k3s uninstaller command: %w", err)
	}
	if uninstallerCmd.ExitCode != 0 {
		return fmt.Errorf("Invalid exit code from k3s uninstaller command: %d", uninstallerCmd.ExitCode)
	}

	return nil
}
