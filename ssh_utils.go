package gconn

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

func sshConfigWithAgentAuth(user string) *ssh.ClientConfig {
	cfg := new(ssh.ClientConfig)
	cfg.User = user
	cfg.HostKeyCallback = checkKnownHosts

	agent := sshAgent()
	if agent != nil {
		signers, err := agent.Signers()
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to get signers from SSH agent: %s", err)
		} else if len(signers) > 0 {
			cfg.Auth = []ssh.AuthMethod{ssh.PublicKeys(signers...)}
		}
	}
	return cfg
}

func sshAgent() agent.Agent {
	sshAuthSocket := os.Getenv("SSH_AUTH_SOCK")
	if sshAuthSocket == "" {
		return nil
	}

	agentConn, err := net.Dial("unix", sshAuthSocket)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to connect to SSH agent: %s", err)
		return nil
	}

	return agent.NewClient(agentConn)
}

func checkKnownHosts(hostname string, remote net.Addr, key ssh.PublicKey) error {
	file, err := os.Open(filepath.Join(os.Getenv("HOME"), ".ssh", "known_hosts"))
	if err != nil {
		return errors.Wrap(err, "failed reading known_hosts file")
	}
	defer file.Close()

	ip := strings.Split(hostname, ":")[0]

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 {
			continue
		}
		if strings.Contains(fields[0], ip) {
			var err error
			hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
			if err != nil {
				return errors.Wrapf(err, "error parsing %q: %v", fields[2], err)
			}
			if bytes.Equal(key.Marshal(), hostKey.Marshal()) {
				return nil
			}
		}
	}

	return errors.Errorf("unknown host (add the following to ~/.ssh/known_hosts)\n%s", MarshalKnownHostEntry(ip, key))
}

func MarshalKnownHostEntry(ip string, key ssh.PublicKey) []byte {
	b := &bytes.Buffer{}
	b.WriteString(ip)
	b.WriteByte(' ')
	b.WriteString(key.Type())
	b.WriteByte(' ')
	e := base64.NewEncoder(base64.StdEncoding, b)
	e.Write(key.Marshal())
	e.Close()
	b.WriteByte('\n')
	return b.Bytes()
}
