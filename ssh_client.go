package gconn

import (
	"fmt"
	"net"
	"os"

	"github.com/pkg/errors"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type sshClient struct {
	addr   string
	client *ssh.Client
	config *ssh.ClientConfig
}

func NewSSHClient(addr, user string) (Client, error) {
	sc := new(sshClient)
	sc.addr = addr
	sc.config = sshConfigWithAgentAuth(user)
	sc.config.Auth = append(sc.config.Auth, ssh.PasswordCallback(sc.askForPassword))

	var err error
	addr = fmt.Sprintf("%s:%d", sc.addr, 22)
	sc.client, err = ssh.Dial("tcp", addr, sc.config)
	return sc, errors.Wrap(err, "failed to connect to SSH host")
}

func (sc *sshClient) NewSession(cmd string, args ...string) (Session, error) {
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

func (sc *sshClient) askForPassword() (string, error) {
	fmt.Printf("Password for %s@%s: ", sc.config.User, sc.addr)
	buf, err := terminal.ReadPassword(int(os.Stdin.Fd()))
	fmt.Printf("\n")
	return string(buf), errors.Wrap(err, "failed to read password")
}
