package gconn

import (
	"net"

	"github.com/pkg/errors"

	"golang.org/x/crypto/ssh"
)

type sshClient struct {
	client *ssh.Client
	config *ssh.ClientConfig
}

// NewSSHClient will connect via SSH to the given address using the user
// specified.
func NewSSHClient(addr, user string, opts ...SSHOption) (Client, error) {
	sCfg, err := newSSHCfg(addr, user, opts)
	if err != nil {
		return nil, err
	}

	sc := &sshClient{config: sCfg.cc}
	sc.client, err = ssh.Dial("tcp", sCfg.Address(), sCfg.cc)
	return sc, errors.Wrap(err, "failed to connect to SSH host")
}

func (sc *sshClient) NewSession(cmd string, args ...string) (Session, error) {
	if sc.client == nil {
		return nil, errors.Errorf("sc.client is nil, why???")
	}
	s, err := sc.client.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new SSH session")
	}

	return &sshSession{Session: s, withSudo: sc.config.User != "root", cmd: cmd, args: args}, nil
}

func (sc *sshClient) Dial(n, addr string) (net.Conn, error) {
	return sc.client.Dial(n, addr)
}

func (sc *sshClient) Close() error {
	return sc.client.Close()
}
