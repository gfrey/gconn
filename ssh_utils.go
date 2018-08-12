package gconn

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
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

	return errors.Errorf("unknown host (add the following to ~/.ssh/known_hosts)\n%s", marshalKnownHostEntry(ip, key))
}

func marshalKnownHostEntry(ip string, key ssh.PublicKey) []byte {
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
