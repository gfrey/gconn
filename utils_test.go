package gconn

import (
	"os/exec"
	"testing"

	"github.com/jgroeneveld/trial/assert"
)

func TestRunSilentSuccess(t *testing.T) {
	cl := NewLocalClient()
	assert.Nil(t, RunSilent(cl, "echo", "foo"))
}

func TestRunSilentError(t *testing.T) {
	cl := NewLocalClient()
	err, ok := RunSilent(cl, "false").(*exec.ExitError)
	assert.True(t, ok)
	assert.Equal(t, 1, err.ExitCode())
}

func TestRunSilentErrorCheckExitCode(t *testing.T) {
	cl := NewLocalClient()
	assert.Equal(t, 55, RunSilent(cl, "bash", "-c", "exit 55").(*exec.ExitError).ExitCode())
	assert.Equal(t, 73, RunSilent(cl, "bash", "-c", "exit 73").(*exec.ExitError).ExitCode())
}

func TestRunStdout(t *testing.T) {
	cl := NewLocalClient()
	stdout, stderr, err := Run(cl, "echo", "foobar")
	assert.MustBeNil(t, err)
	assert.Equal(t, "foobar\n", stdout)
	assert.Equal(t, "", stderr)
}

func TestRunStderr(t *testing.T) {
	cl := NewLocalClient()
	stdout, stderr, err := Run(cl, "bash", "-c", "echo foobar >&2")
	assert.MustBeNil(t, err)
	assert.Equal(t, "", stdout)
	assert.Equal(t, "foobar\n", stderr)
}

func TestRunError(t *testing.T) {
	cl := NewLocalClient()
	stdout, stderr, err := Run(cl, "bash", "-c", "echo foobar >&2 && exit 1")
	assert.Equal(t, 1, err.(*exec.ExitError).ExitCode())
	assert.Equal(t, "", stdout)
	assert.Equal(t, "foobar\n", stderr)
}
