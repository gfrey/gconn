package gconn

import (
	"io"
	"os/exec"
)

type localClient struct{}

func NewLocalClient() Client {
	return &localClient{}
}

func (lc *localClient) Close() error {
	// TODO need to wait for all sessions to close
	return nil
}

func (lc *localClient) NewSession(cmd string, args ...string) (Session, error) {
	return newLocalSession(cmd, args)
}

type localSession struct {
	cmd      *exec.Cmd
	finished bool
}

func newLocalSession(cmd string, args []string) (Session, error) {
	return &localSession{cmd: exec.Command(cmd, args...)}, nil
}

func (ls *localSession) Close() error {
	if !ls.finished {
		return ls.cmd.Wait()
	}
	return nil
}

func (ls *localSession) Run() error {
	ls.finished = true
	return ls.cmd.Run()
}

func (ls *localSession) Start() error {
	return ls.cmd.Start()
}
func (ls *localSession) Wait() error {
	ls.finished = true
	return ls.cmd.Wait()
}

func (ls *localSession) StdinPipe() (io.WriteCloser, error) {
	return ls.cmd.StdinPipe()
}
func (ls *localSession) StdoutPipe() (io.Reader, error) {
	return ls.cmd.StdoutPipe()
}
func (ls *localSession) StderrPipe() (io.Reader, error) {
	return ls.cmd.StderrPipe()
}
