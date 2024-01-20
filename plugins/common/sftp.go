package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/melbahja/goph"
	"golang.org/x/crypto/ssh"
)

// SftpCopyInput is the input for CallSftpCopy
type SftpCopyInput struct {
	// AllowUknownHosts allows connecting to hosts with unknown host keys
	AllowUknownHosts bool

	// DestinationPath is the path to copy the file to
	DestinationPath string

	// RemoteHost is the remote host to connect to
	RemoteHost string

	// SourcePath is the path to the file to copy
	SourcePath string
}

// CallSftpCopy copies a file to a remote host via sftp
func CallSftpCopy(input SftpCopyInput) (SftpCopyResult, error) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
		syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-signals
		cancel()
	}()

	u, err := url.Parse(input.RemoteHost)
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to parse remote host: %w", err)
	}
	if u.Scheme == "" {
		return SftpCopyResult{}, fmt.Errorf("missing remote host ssh scheme in remote host: %s", input.RemoteHost)
	}
	if u.Scheme != "ssh" {
		return SftpCopyResult{}, fmt.Errorf("invalid remote host scheme: %s", u.Scheme)
	}

	username := ""
	password := ""
	if u.User != nil {
		username = u.User.Username()
		if pass, ok := u.User.Password(); ok {
			password = pass
		}
	}
	if username == "" {
		username = os.Getenv("USER")
	}

	portStr := u.Port()
	port := 0
	if portStr != "" {
		portVal, err := strconv.Atoi(portStr)
		if err != nil {
			return SftpCopyResult{}, fmt.Errorf("failed to parse port: %w", err)
		}
		port = portVal
	}
	if port == 0 {
		port = 22
	}

	sshKeyPath := filepath.Join(os.Getenv("DOKKU_ROOT"), ".ssh/id_ed25519")
	if !FileExists(sshKeyPath) {
		sshKeyPath = filepath.Join(os.Getenv("DOKKU_ROOT"), ".ssh/id_rsa")
	}
	if !FileExists(sshKeyPath) {
		return SftpCopyResult{}, errors.New("ssh key not found at ~/.ssh/id_ed25519 or ~/.ssh/id_rsa")
	}

	cmd := SftpCopyTask{
		SourcePath:       input.SourcePath,
		DestinationPath:  input.DestinationPath,
		AllowUknownHosts: input.AllowUknownHosts,
		Hostname:         u.Hostname(),
		Port:             uint(port),
		Username:         username,
		Password:         password,
		SshKeyPath:       sshKeyPath,
	}

	if os.Getenv("DOKKU_TRACE") == "1" {
		cmd.PrintCommand = true
	}

	res, err := cmd.Execute(ctx)
	if err != nil {
		return res, err
	}

	return res, nil
}

type SftpCopyTask struct {
	// SourcePath is the path to the file to copy
	SourcePath string

	// DestinationPath is the path to copy the file to
	DestinationPath string

	// Shell run the command in a bash shell.
	// Note that the system must have `bash` installed in the PATH or in /bin/bash
	Shell bool

	// Stdin connect a reader to stdin for the command
	// being executed.
	Stdin io.Reader

	// PrintCommand prints the command before executing
	PrintCommand bool

	// AllowUknownHosts allows connecting to hosts with unknown host keys
	AllowUknownHosts bool

	// Hostname is the hostname to connect to
	Hostname string

	// Port is the port to connect to
	Port uint

	// Username is the username to connect with
	Username string

	// Password is the password to connect with
	Password string

	// SshKeyPath is the path to the ssh key to use
	SshKeyPath string
}

// SftpCopyResult is the result of executing an SftpCopyTask
type SftpCopyResult struct {
	ExitErr   error
	Cancelled bool
}

// Execute runs the task
func (task SftpCopyTask) Execute(ctx context.Context) (SftpCopyResult, error) {
	if task.SourcePath == "" {
		return SftpCopyResult{}, errors.New("source path is required")
	}
	if task.DestinationPath == "" {
		return SftpCopyResult{}, errors.New("destination path is required")
	}
	if task.Hostname == "" {
		return SftpCopyResult{}, errors.New("hostname is required")
	}
	if task.SshKeyPath == "" {
		return SftpCopyResult{}, errors.New("ssh key path is required")
	}
	if task.Username == "" {
		return SftpCopyResult{}, errors.New("username is required")
	}
	if task.Port == 0 {
		task.Port = 22
	}

	auth, err := goph.Key(task.SshKeyPath, "")
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to load ssh key: %w", err)
	}

	callback, err := goph.DefaultKnownHosts()
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to load known hosts: %w", err)
	}

	if task.AllowUknownHosts {
		callback = ssh.InsecureIgnoreHostKey()
	}

	connectionConf := goph.Config{
		User:     task.Username,
		Addr:     task.Hostname,
		Port:     task.Port,
		Timeout:  goph.DefaultTimeout,
		Callback: callback,
		Auth:     auth,
	}
	if task.Password != "" {
		connectionConf.Auth = goph.Password(task.Password)
	}

	client, err := goph.NewConn(&connectionConf)
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to create ssh client: %w", err)
	}
	defer client.Close()

	// don't try to run if the context is already cancelled
	if ctx.Err() != nil {
		return SftpCopyResult{
			Cancelled: ctx.Err() == context.Canceled,
		}, ctx.Err()
	}

	if task.PrintCommand {
		LogDebug(fmt.Sprintf("ssh %s@%s cp %s %v", task.Username, task.Hostname, task.SourcePath, task.DestinationPath))
	}

	sftp, err := client.NewSftp()
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to create sftp client: %w", err)
	}

	contents, err := os.ReadFile(task.SourcePath)
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to read source file: %w", err)
	}

	dstFile, err := sftp.Create(task.DestinationPath)
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to create destination file: %w", err)
	}

	_, err = dstFile.Write(contents)
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to write to destination file: %w", err)
	}

	err = dstFile.Close()
	if err != nil {
		return SftpCopyResult{}, fmt.Errorf("failed to create ssh command: %w", err)
	}

	err = sftp.Close()

	return SftpCopyResult{
		ExitErr:   err,
		Cancelled: ctx.Err() == context.Canceled,
	}, ctx.Err()
}
