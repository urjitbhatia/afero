package sshfs_test

import (
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/pkg/sftp"
)

const (
	READONLY                = true
	READWRITE               = false
	NO_DELAY  time.Duration = 0

	debuglevel = "ERROR" // set to "DEBUG" for debugging
)

type delayedWrite struct {
	t time.Time
	b []byte
}

func testClientSvr(t testing.TB) *sftp.Client {
	c1, c2 := netPipe(t)

	options := []sftp.ServerOption{sftp.WithDebug(os.Stderr)}

	server, err := sftp.NewServer(c1, options...)
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve()

	var ctx io.WriteCloser = c2

	client, err := sftp.NewClientPipe(c2, ctx)
	if err != nil {
		t.Fatal(err)
	}

	return client
}

// netPipe provides a pair of io.ReadWriteClosers connected to each other.
// The functions is identical to os.Pipe with the exception that netPipe
// provides the Read/Close guarantees that os.File derrived pipes do not.
func netPipe(t testing.TB) (io.ReadWriteCloser, io.ReadWriteCloser) {
	type result struct {
		net.Conn
		error
	}

	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	ch := make(chan result, 1)
	go func() {
		conn, err := l.Accept()
		ch <- result{conn, err}
		err = l.Close()
		if err != nil {
			t.Error(err)
		}
	}()
	c1, err := net.Dial("tcp", l.Addr().String())
	if err != nil {
		l.Close() // might cause another in the listening goroutine, but too bad
		t.Fatal(err)
	}
	r := <-ch
	if r.error != nil {
		t.Fatal(err)
	}
	return c1, r.Conn
}
