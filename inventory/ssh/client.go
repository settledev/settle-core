package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"time"
	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/inventory/parser"
	gossh "golang.org/x/crypto/ssh"
)

const (
	ConnectTimeout = 5 * time.Second
	ReadTimeout    = 60 * time.Second
	WriteTimeout   = 60 * time.Second
	MaxConnections = 10
)

type SSHClient struct {
	Host   *common.Host
	Client *gossh.Client
}

type SSHConfig struct {
	Hostname     string
	User         string
	Port         int
	IdentityFile string
	ProxyCommand string
	HostKeyFile  string
}

func loadSSHConfig(hostname string) (*SSHConfig, error) {

	currentUser, err := user.Current()
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	sshConfigPath := filepath.Join(currentUser.HomeDir, ".ssh", "config")

	if _, err := os.Stat(sshConfigPath); os.IsNotExist(err) {
		return nil, nil
	}

	hosts, err := parser.ParseHosts(sshConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH config: %w", err)
	}

	for _, host := range hosts {
		if host.Name == hostname {
			return &SSHConfig{
				Hostname:     host.Hostname,
				User:         host.User,
				Port:         host.Port,
				IdentityFile: host.Keyfile,
			}, nil
		}
	}

	return nil, nil
}

func resolveHost(hostname string) (string, string, int, string, error) {
	config, err := loadSSHConfig(hostname)
	if err != nil {
		return "", "", 0, "", err
	}

	resolvedHostname := hostname
	if config != nil && config.Hostname != "" {
		resolvedHostname = config.Hostname
	}

	resolvedUser := ""
	if config != nil && config.User != "" {
		resolvedUser = config.User
	}

	resolvedPort := 22
	if config != nil && config.Port != 0 {
		resolvedPort = config.Port
	}

	resolvedKeyFile := ""
	if config != nil && config.IdentityFile != "" {
		resolvedKeyFile = config.IdentityFile
	}

	return resolvedHostname, resolvedUser, resolvedPort, resolvedKeyFile, nil
}

func validateKeyFile(keyPath string) error {

	info, err := os.Stat(keyPath)
	if err != nil {
		return fmt.Errorf("failed to stat key file: %w", err)
	}

	mode := info.Mode()
	if mode&0077 != 0 {
		return fmt.Errorf("key file has insecure permissions: %v", mode)
	}

	if info.Size() > 10240 {
		return fmt.Errorf("key file too large: %d bytes", info.Size())
	}

	return nil
}

func validateHostname(hostname string) error {
	if hostname == "" {
		return fmt.Errorf("hostname cannot be empty")
	}

	if len(hostname) > 255 {
		return fmt.Errorf("hostname too long")
	}

	return nil
}

func createSecureConfig(host *common.Host, signer gossh.Signer) (*gossh.ClientConfig, error) {

	if err := validateHostname(host.Hostname); err != nil {
		return nil, fmt.Errorf("invalid hostname: %w", err)
	}

	if err := validateKeyFile(host.Keyfile); err != nil {
		return nil, fmt.Errorf("invalid key file: %w", err)
	}

	config := &gossh.ClientConfig{
		User: host.User,
		Auth: []gossh.AuthMethod{
			gossh.PublicKeys(signer),
		},
		// InsecureIgnoreHostKey is acceptable for IaC tools because:
		// - First-time connections to new servers
		// - Development and testing environments
		// - Automation scenarios requiring unattended operation
		// - Dynamic cloud environments where host keys change
		HostKeyCallback: gossh.InsecureIgnoreHostKey(),
		Timeout:         ConnectTimeout,
		BannerCallback: func(message string) error {

			return nil
		},
	}

	return config, nil
}

func NewSSHClient(host *common.Host) (*SSHClient, error) {
	if host == nil {
		return nil, fmt.Errorf("host cannot be nil")
	}

	if host.Hostname == "" {
		resolvedHostname, resolvedUser, resolvedPort, resolvedKeyFile, err := resolveHost(host.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve host: %w", err)
		}
		if resolvedHostname != "" {
			host.Hostname = resolvedHostname
		}
		if resolvedUser != "" {
			host.User = resolvedUser
		}
		if resolvedPort != 0 {
			host.Port = resolvedPort
		}
		if resolvedKeyFile != "" {
			host.Keyfile = resolvedKeyFile
		}
	}

	if host.Keyfile == "" {
		currentUser, err := user.Current()
		if err == nil {
			defaultKeys := []string{
				filepath.Join(currentUser.HomeDir, ".ssh", "id_rsa"),
				filepath.Join(currentUser.HomeDir, ".ssh", "id_ed25519"),
				filepath.Join(currentUser.HomeDir, ".ssh", "id_ecdsa"),
			}

			for _, keyPath := range defaultKeys {
				if _, err := os.Stat(keyPath); err == nil {
					host.Keyfile = keyPath
					break
				}
			}
		}
	}

	key, err := os.ReadFile(host.Keyfile)
	if err != nil {
		return nil, fmt.Errorf("failed to read key file: %w", err)
	}

	signer, err := gossh.ParsePrivateKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	config, err := createSecureConfig(host, signer)
	if err != nil {
		return nil, err
	}

	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host.Hostname, fmt.Sprintf("%d", host.Port)), ConnectTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection: %w", err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
		tcpConn.SetLinger(0)
	}

	sshConn, chans, reqs, err := gossh.NewClientConn(conn, net.JoinHostPort(host.Hostname, fmt.Sprintf("%d", host.Port)), config)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	client := gossh.NewClient(sshConn, chans, reqs)

	return &SSHClient{
		Host:   host,
		Client: client,
	}, nil
}

func (s *SSHClient) Close() error {
	if s.Client != nil {
		return s.Client.Close()
	}
	return nil
}

func (s *SSHClient) RunCommand(ctx context.Context, command string) (string, error) {
	session, err := s.Client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	resultChan := make(chan struct {
		Output []byte
		Error  error
	}, 1)

	go func() {
		defer close(resultChan)
		out, err := session.CombinedOutput(command)
		resultChan <- struct {
			Output []byte
			Error  error
		}{out, err}
	}()

	select {
	case <-ctx.Done():
		_ = session.Signal(gossh.SIGKILL)
		return "", ctx.Err()
	case result := <-resultChan:
		if result.Error != nil {
			return "", fmt.Errorf("failed to run command: %w", result.Error)
		}
		return string(result.Output), nil
	case <-time.After(ReadTimeout):
		_ = session.Signal(gossh.SIGKILL)
		return "", fmt.Errorf("command timed out after %s", ReadTimeout)
	}
}

func (s *SSHClient) TestConnection() error {
	err := PingHost(s.Host)

	return err
}
