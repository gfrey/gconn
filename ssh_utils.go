package gconn

import (
	"net"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
)

func sshAgent() (agent.Agent, error) {
	sshAuthSocket := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSocket == "" {
		return nil, errors.Errorf("env variable SSH_AUTH_SOCK not set")
	}

	agentConn, err := net.Dial("unix", sshAuthSocket)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to SSH agent")
	}

	return agent.NewClient(agentConn), nil
}

func checkKnownHosts(hostname string, remote net.Addr, key ssh.PublicKey) error {
	khPath := filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts")
	callback, err := knownhosts.New(khPath)
	if err != nil {
		return errors.Wrap(err, "failed reading known_hosts file")
	}
	return callback(hostname, remote, key)
}
