package inventory

import (
	"fmt"

	"github.com/settlectl/settle-core/common"
	"github.com/settlectl/settle-core/inventory/ssh"
)

type Context struct {
	Host      *common.Host
	SSHClient *ssh.SSHClient
	Logger    *Logger
}

func NewContext(host *common.Host) *Context {
	return &Context{
		Host:      host,
		SSHClient: nil,
		Logger:    NewLogger(),
	}
}

// CreateSSHClient creates an SSH client for the given host
func (c *Context) CreateSSHClient(host *common.Host) (*ssh.SSHClient, error) {
	sshClient, err := ssh.NewSSHClient(host)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH client: %w", err)
	}
	return sshClient, nil
}

// SetHost sets the host for this context
func (c *Context) SetHost(host *common.Host) {
	c.Host = host
}

// SetSSHClient sets the SSH client for this context
func (c *Context) SetSSHClient(client *ssh.SSHClient) {
	c.SSHClient = client
}
