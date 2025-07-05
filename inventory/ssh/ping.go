package ssh

import (
	"github.com/settlectl/settle-core/common"
)

func PingHost(host *common.Host) error {
	client, err := NewSSHClient(host)
	if err != nil {
		return err
	}
	defer client.Close()

	return nil
}