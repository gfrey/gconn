package gconn

import (
	"bufio"
	"bytes"
	"io"
	"log"

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
