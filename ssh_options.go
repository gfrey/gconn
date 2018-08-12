package gconn

import (
	"fmt"
	"net"
	"os"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type sshCfg struct {
	cc   *ssh.ClientConfig
	addr string
	user string
	port int
}

func newSSHCfg(addr, user string, opts []SSHOption) (*sshCfg, error) {
	sc := &sshCfg{addr: addr, user: user, port: 22}
	sc.cc = &ssh.ClientConfig{User: user}

	if err := sc.applyOptions(opts...); err != nil {
		return nil, err
	}

	// steps required to make it behave as before
	defaultOpts := []SSHOption{}
	if sc.cc.HostKeyCallback == nil {
		defaultOpts = append(defaultOpts, SSHWithHostCheckFile())
	}

	if len(sc.cc.Auth) == 0 {
		defaultOpts = append(defaultOpts, SSHWithAgent(), SSHWithAskPassword())
	}

	return sc, sc.applyOptions(defaultOpts...)
}

func (sc *sshCfg) applyOptions(opts ...SSHOption) error {
	for i := range opts {
		if err := opts[i](sc); err != nil {
			return err
		}
	}
	return nil
}

func (sc *sshCfg) Address() string {
	return fmt.Sprintf("%s:%d", sc.addr, sc.port)
}

type SSHOption func(*sshCfg) error

func SSHWithPort(port int) SSHOption {
	return func(cfg *sshCfg) error {
		cfg.port = port
		return nil
	}
}

func SSHWithHostCheckFile() SSHOption {
	return func(cfg *sshCfg) error {
		cfg.cc.HostKeyCallback = checkKnownHosts
		return nil
	}
}

func SSHWithHostCheck(f func(string, net.Addr, ssh.PublicKey) error) SSHOption {
	return func(cfg *sshCfg) error {
		cfg.cc.HostKeyCallback = f
		return nil
	}
}

func SSHWithAgent() SSHOption {
	return func(cfg *sshCfg) error {
		agent, err := sshAgent()
		if err != nil {
			return err
		}

		signers, err := agent.Signers()
		if err != nil {
			return errors.Wrap(err, "failed to get signers from SSH agent")
		}
		cfg.cc.Auth = append(cfg.cc.Auth, ssh.PublicKeys(signers...))
		return nil
	}
}

func SSHWithAskPassword() SSHOption {
	return func(cfg *sshCfg) error {
		pwdC := func() (string, error) {
			fmt.Printf("Password for %s@%s: ", cfg.user, cfg.addr)
			buf, err := terminal.ReadPassword(int(os.Stdin.Fd()))
			fmt.Printf("\n")
			return string(buf), errors.Wrap(err, "failed to read password")
		}
		cfg.cc.Auth = append(cfg.cc.Auth, ssh.PasswordCallback(pwdC))
		return nil
	}
}

func SSHWithCustomAskPassword(f func() (string, error)) SSHOption {
	return func(cfg *sshCfg) error {
		cfg.cc.Auth = append(cfg.cc.Auth, ssh.PasswordCallback(f))
		return nil
	}
}
