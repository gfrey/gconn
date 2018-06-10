package gconn

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// NewSSHProxyClient will create a connection through a proxy client, i.e. the
// given client will be used as a proxy to the target configured in this
// client.
func NewSSHProxyClient(proxy Client, addr, user string) (Client, error) {
	sc := new(sshClient)
	sc.addr = fmt.Sprintf("%s:%d", addr, 22)

	conn, err := proxy.Dial("tcp", sc.addr)
	if err != nil {
		return nil, err
	}

	sc.config = sshConfigWithAgentAuth(user)
	ncc, chans, reqs, err := ssh.NewClientConn(conn, sc.addr, sc.config)
	if err != nil {
		return nil, err
	}
	sc.client = ssh.NewClient(ncc, chans, reqs)
	return sc, nil
}
