package gconn

import (
	"bufio"
	"io"

	"github.com/gfrey/glog"
	"golang.org/x/sync/errgroup"
)

// Run a command with the given client ignoring all output.
func Run(c Client, cmd string, args ...string) error {
	sess, err := c.NewSession(cmd, args...)
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.Run()
}

// Run a command with the given client sending all output to tagged logger.
func RunWithLogger(l glog.Logger, c Client, cmd string, args ...string) error {
	sess, err := c.NewSession(cmd, args...)
	if err != nil {
		return err
	}
	defer sess.Close()

	stdout, err := sess.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := sess.StderrPipe()
	if err != nil {
		return err
	}

	g := errgroup.Group{}
	g.Go(readStream(l.Tag("stdout"), stdout))
	g.Go(readStream(l.Tag("stderr"), stderr))
	g.Go(func() error { return sess.Run() })

	return g.Wait()
}

func readStream(l glog.Logger, r io.Reader) func() error {
	return func() error {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			l.Printf("%s", sc.Text())
		}
		return sc.Err()
	}
}
