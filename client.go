package gconn

import (
	"io"
	"net"
)

// Session provides the interface for a command execution session.
type Session interface {
	// Close the session. Might return io.EOF if the session was already
	// closed. Especially with SSH that is expected behavior.
	Close() error

	// Get a handle to the standard input of the command session.
	StdinPipe() (io.WriteCloser, error)
	// Get a handle to the standard output of the command session.
	StdoutPipe() (io.Reader, error)
	// Get a handle to the standard error output of the command session.
	StderrPipe() (io.Reader, error)

	// Run the session. This will trigger both `Start` and `Wait`.
	Run() error
	// Start the session.
	Start() error
	// Wait for the session to finish.
	Wait() error
}

// Client will create provide the connection to the target and the session for
// commands to be executed.
type Client interface {
	// Create a new session for the given command.
	NewSession(cmd string, args ...string) (Session, error)
	// Dial to the given address and get a connection.
	Dial(n, addr string) (net.Conn, error)
	// Close the client and all respective sessions.
	Close() error
}
