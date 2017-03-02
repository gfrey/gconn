package gconn

import (
	"bufio"
	"io"
	"sync"

	"github.com/gfrey/glog"
	"github.com/pkg/errors"
)

type loggedClient struct {
	Client

	l glog.Logger
}

func NewLoggedClient(l glog.Logger, c Client) Client {
	return &loggedClient{Client: c, l: l}
}

func (lc *loggedClient) NewSession(cmd string, args ...string) (Session, error) {
	sess, err := lc.Client.NewSession(cmd, args...)
	if err != nil {
		return nil, err
	}

	return newLoggedSession(lc.l, sess)
}

type loggedSession struct {
	Session

	wg *sync.WaitGroup
}

func newLoggedSession(l glog.Logger, sess Session) (Session, error) {
	s := &loggedSession{Session: sess, wg: new(sync.WaitGroup)}

	stdout, err := s.Session.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build stdout pipe")
	}

	stderr, err := s.Session.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build stderr pipe")
	}

	s.wg.Add(2)
	go s.readStream(l.Tag("stdout"), stdout)
	go s.readStream(l.Tag("stderr"), stderr)

	return s, nil
}

func (lsess *loggedSession) Close() error {
	err := lsess.Session.Close()
	lsess.wg.Wait()
	return err
}

func (lsess *loggedSession) StdoutPipe() (io.Reader, error) {
	return nil, errors.New("logged session has no access to stdout pipe!")
}

func (lsess *loggedSession) StderrPipe() (io.Reader, error) {
	return nil, errors.New("logged session has no access to stderr pipe!")
}

func (lsess *loggedSession) readStream(l glog.Logger, stream io.Reader) {
	defer lsess.wg.Done()

	sc := bufio.NewScanner(stream)
	for sc.Scan() {
		l.Printf(sc.Text())
	}

	if err := sc.Err(); err != nil {
		l.Printf("failed scanning stderr: %s", err)
	}
}