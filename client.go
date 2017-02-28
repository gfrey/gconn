package gconn

import "io"

type Session interface {
	Close() error

	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.Reader, error)
	StderrPipe() (io.Reader, error)

	Run() error
	Start() error
	Wait() error
}

type Client interface {
	NewSession(cmd string, args ...string) (Session, error)
	Close() error
}
