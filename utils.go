package gconn

import (
	"bufio"
	"bytes"
	"io"
	"log"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// Run a command with the given client ignoring all output.
func RunSilent(c Client, cmd string, args ...string) error {
	sess, err := c.NewSession(cmd, args...)
	if err != nil {
		return err
	}
	defer sess.Close()

	return sess.Run()
}

func Run(c Client, cmd string, args ...string) (string, string, error) {
	sess, err := c.NewSession(cmd, args...)
	if err != nil {
		return "", "", err
	}
	defer sess.Close()

	stdout, err := sess.StdoutPipe()
	if err != nil {
		return "", "", err
	}
	bufStdout, captureFnStdout := captureStream(stdout)

	stderr, err := sess.StderrPipe()
	if err != nil {
		return "", "", err
	}
	bufStderr, captureFnStderr := captureStream(stderr)

	g := errgroup.Group{}
	g.Go(captureFnStdout)
	g.Go(captureFnStderr)
	g.Go(func() error { return sess.Run() })
	err = g.Wait()
	return bufStdout.String(), bufStderr.String(), err
}

// The callback is given a line of output and returns the item search. The
// boolean indicates whether to keep scanning.
func RunWithScanner(c Client, fn func(string) (string, bool), cmd string, args ...string) (string, error) {
	// determine whether the VM in question already exists
	sess, err := c.NewSession(cmd, args...)
	if err != nil {
		return "", err
	}
	defer sess.Close()

	stdout, err := sess.StdoutPipe()
	if err != nil {
		return "", err
	}
	if err := sess.Start(); err != nil {
		return "", err
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, stdout); err != nil {
		return "", err
	}

	if err := sess.Wait(); err != nil {
		return "", err
	}

	result := ""

	sc := bufio.NewScanner(buf)
	for sc.Scan() {
		r, stop := fn(sc.Text())
		if r != "" {
			result = r
		}
		if stop {
			return result, nil
		}
	}

	if err := sc.Err(); err != nil {
		return "", errors.Wrap(err, "failed to scan output")
	}

	return result, nil
}

// RunWithLogger will send all output of the command to the given logger.
func RunWithLogger(l *log.Logger, c Client, cmd string, args ...string) error {
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

	return runSessionLoggingStreams(l, sess, stdout, stderr)
}

func runSessionLoggingStreams(l *log.Logger, sess Session, stdout, stderr io.Reader) error {
	g := errgroup.Group{}
	g.Go(readStream(l, "stdout", stdout))
	g.Go(readStream(l, "stderr", stderr))
	g.Go(func() error { return sess.Run() })

	return g.Wait()
}

func readStream(l *log.Logger, s string, r io.Reader) func() error {
	return func() error {
		sc := bufio.NewScanner(r)
		for sc.Scan() {
			l.Printf("%s %s", s, sc.Text())
		}
		return sc.Err()
	}
}

func captureStream(r io.Reader) (*bytes.Buffer, func() error) {
	buf := bytes.NewBuffer(nil)
	fn := func() error {
		_, err := io.Copy(buf, r)
		return err
	}
	return buf, fn
}
