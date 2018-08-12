package gconn

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
	"testing"

	"github.com/gfrey/sshtest"
	"github.com/jgroeneveld/trial/assert"
)

func TestSSHWrite(t *testing.T) {
	ss, err := sshtest.NewServer(sshtest.WithPasswordCallbackForUser("root", "the-secret"))
	assert.Nil(t, err, "didn't expect an error")
	assert.MustBeNil(t, ss.Start(), "didn't expect an error")

	ss.RegisterHandler("test_write", handleTestWrite)

	cl, err := NewSSHClient(ss.Addr, "root",
		SSHWithPort(ss.Port),
		SSHWithHostCheck(ss.CheckHostKey),
		SSHWithCustomAskPassword(func() (string, error) { return "the-secret", nil }),
	)
	assert.MustBeNil(t, err, "didn't expect an error creating the client")

	for i := 1; i < 18000; i = i * 2 {
		se, err := cl.NewSession("test_write", strconv.Itoa(i))
		assert.Nil(t, err, "didn't expect an error")

		stdout, err := se.StdoutPipe()
		assert.MustBeNil(t, err, "didn't expect an error receiving the stdout pipe")

		assert.MustBeNil(t, se.Start(), "didn't expect an error starting the session")

		buf := bytes.NewBuffer(nil)
		n, err := io.Copy(buf, stdout)
		assert.MustBeNil(t, err, "didn't expect an error copying data to a buffer")
		assert.Equal(t, int64(i), n, "exected %d bytes", i)

		assert.Nil(t, se.Wait(), "didn't expect an error waiting for the session to end")
		assert.Equal(t, io.EOF, se.Close(), "expected EOF error, as session should be closed by workflow")

		assert.Equal(t, i, buf.Len(), "expecte %d bytes", i)
		for j := 0; j < i; j++ {
			assert.Equal(t, byte(j%256), buf.Bytes()[j], "byte %d has unexpected value", j)
		}

	}
	assert.Nil(t, ss.Close(), "didn't expect an error closing the server")
}

func TestSSHRead(t *testing.T) {
	ss, err := sshtest.NewServer(sshtest.WithPasswordCallbackForUser("root", "the-secret"))
	assert.Nil(t, err, "didn't expect an error")
	assert.MustBeNil(t, ss.Start(), "didn't expect an error")

	ss.RegisterHandler("test_read", handleTestRead)

	cl, err := NewSSHClient(ss.Addr, "root",
		SSHWithPort(ss.Port),
		SSHWithHostCheck(ss.CheckHostKey),
		SSHWithCustomAskPassword(func() (string, error) { return "the-secret", nil }),
	)
	assert.MustBeNil(t, err, "didn't expect an error creating the client")

	for i := 1; i < 18000; i = i * 2 {
		se, err := cl.NewSession("test_read", strconv.Itoa(i))
		assert.Nil(t, err, "didn't expect an error")

		stdin, err := se.StdinPipe()
		assert.MustBeNil(t, err, "didn't expect an error receiving the stdin pipe")

		assert.MustBeNil(t, se.Start(), "didn't expect an error starting the session")

		n, err := stdin.Write(createBytes(i))
		assert.MustBeNil(t, err, "didn't expect an error sending data stdin")
		assert.Equal(t, i, n, "exected %d bytes", i)
		assert.MustBeNil(t, stdin.Close(), "didn't expect an error closing stdin")

		assert.Nil(t, se.Wait(), "didn't expect an error waiting for the session to end")
		assert.Equal(t, io.EOF, se.Close(), "expected EOF error, as session should be closed by workflow")
	}
	assert.Nil(t, ss.Close(), "didn't expect an error closing the server")
}

func handleTestWrite(cmd string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	switch len(args) {
	case 0:
		io.WriteString(stderr, "err: missing length of string\n")
		return 1
	case 1:
	default:
		io.WriteString(stderr, "err: wrong number of args\n")
		return 1
	}

	l, err := strconv.Atoi(args[0])
	if err != nil {
		io.WriteString(stderr, err.Error())
		return 1
	}

	stdout.Write(createBytes(l))
	return 0
}

func handleTestRead(cmd string, args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	switch len(args) {
	case 0:
		io.WriteString(stderr, "err: missing length of string\n")
		return 1
	case 1:
	default:
		io.WriteString(stderr, "err: wrong number of args\n")
		return 1
	}

	l, err := strconv.Atoi(args[0])
	if err != nil {
		io.WriteString(stderr, err.Error())
		return 1
	}

	buf := bytes.NewBuffer(nil)
	n, err := io.Copy(buf, stdin)
	if err != nil {
		io.WriteString(stderr, err.Error())
		return 1
	}

	if n != int64(l) {
		io.WriteString(stderr, fmt.Sprintf("exected %d bytes, got %d\n", l, n))
		return 1
	}

	if buf.Len() != l {
		io.WriteString(stderr, fmt.Sprintf("exected %d bytes in buffer, got %d\n", l, buf.Len()))
		return 1
	}

	for i := 0; i < l; i++ {
		if buf.Bytes()[i] != byte(i%256) {
			io.WriteString(stderr, fmt.Sprintf("exected byte %d to be %q, got %q\n", i, string(byte(i%256)), buf.Bytes()[i]))
			return 1
		}
	}

	return 0
}

func createBytes(l int) []byte {
	res := make([]byte, l)
	for i := 0; i < l; i++ {
		res[i] = byte(i % 256)
	}
	return res
}
