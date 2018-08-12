package gconn

import (
	"golang.org/x/crypto/ssh"
)

// NewSSHProxyClient will create a connection through a proxy client, i.e. the
// given client will be used as a proxy to the target configured in this
// client.
func NewSSHProxyClient(proxy Client, addr, user string, opts ...SSHOption) (Client, error) {
	sCfg, err := newSSHCfg(addr, user, opts)
	if err != nil {
		return nil, err
	}

	conn, err := proxy.Dial("tcp", sCfg.Address())
	if err != nil {
		return nil, err
	}

	ncc, chans, reqs, err := ssh.NewClientConn(conn, sCfg.Address(), sCfg.cc)
	if err != nil {
		return nil, err
	}

	sc := &sshClient{config: sCfg.cc}
	sc.client = ssh.NewClient(ncc, chans, reqs)
	return sc, nil
}
