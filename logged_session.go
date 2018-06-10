package gconn

import (
	"bufio"
	"io"
	"log"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type loggedClient struct {
	Client

	l *log.Logger
}

func NewLoggedClient(l *log.Logger, c Client) Client {
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

	errG *errgroup.Group
}

func newLoggedSession(l *log.Logger, sess Session) (Session, error) {
	s := &loggedSession{Session: sess, errG: new(errgroup.Group)}

	stdout, err := s.Session.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build stdout pipe")
	}

	stderr, err := s.Session.StderrPipe()
	if err != nil {
		return nil, errors.Wrap(err, "failed to build stderr pipe")
	}

	s.errG.Go(readStream(l, "stdout", stdout))
	s.errG.Go(readStream(l, "stderr", stderr))

	return s, nil
}

func (lsess *loggedSession) Close() error {
	err := lsess.Session.Close()
	if errStream := lsess.errG.Wait(); errStream != nil {
		err = multierror.Append(err, errStream)
	}
	return err
}

func (lsess *loggedSession) StdoutPipe() (io.Reader, error) {
	return nil, errors.New("logged session has no access to stdout pipe!")
}

func (lsess *loggedSession) StderrPipe() (io.Reader, error) {
	return nil, errors.New("logged session has no access to stderr pipe!")
}

func (lsess *loggedSession) readStream(l *log.Logger, sname string, stream io.Reader) error {
	sc := bufio.NewScanner(stream)
	for sc.Scan() {
		l.Printf(sname + " " + sc.Text())
	}

	return errors.Wrapf(sc.Err(), "failed scanning %s", sname)
}
