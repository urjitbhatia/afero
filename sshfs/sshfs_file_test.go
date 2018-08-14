package sshfs_test

import (
	"io"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/pkg/errors"
)

func TestReadWrite(t *testing.T) {
	testFileName := "testReadWriteFile"
	// Get test file system
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, errors.WithStack(err))
	sub := path.Join(dir, testFileName)

	// file.Create
	rwf, err := fs.Create(sub)
	fatalOnErr(t, errors.WithStack(err))

	// file.WriteString
	_, err = rwf.WriteString("Hello World")
	fatalOnErr(t, errors.WithStack(err))

	// file.Close
	fatalOnErr(t, errors.WithStack(rwf.Close()))

	// fs.Open
	rwf, err = fs.OpenFile(sub, os.O_RDWR, os.ModeAppend)
	fatalOnErr(t, errors.WithStack(err))
	defer rwf.Close()

	// file.WriteAt
	_, err = rwf.WriteAt([]byte(" golang.org"), 11)
	fatalOnErr(t, errors.WithStack(err))

	// file.Seek
	_, err = rwf.Seek(0, io.SeekStart)
	fatalOnErr(t, errors.WithStack(err))

	// file.Read
	buf, err := ioutil.ReadAll(rwf)
	fatalOnErr(t, errors.WithStack(err))

	bufs := string(buf)
	if bufs != "Hello World golang.org" {
		t.Fatalf("Expected file to contain: 'Hello World golang.org' Found: %s", bufs)
	}

	// file.ReadAt
	buf = make([]byte, 5)       // Read World
	_, err = rwf.ReadAt(buf, 6) // Skip 'Hello '
	fatalOnErr(t, errors.WithStack(err))

	bufs = string(buf)
	if bufs != "World" {
		t.Fatalf("Expected file to contain: 'World' Found: %s", bufs)
	}

	// file.Name
	if rwf.Name() != sub {
		t.Fatalf("Expected file name to be: '%s' Found: %s", sub, rwf.Name())
	}

	// file.Stat
	info, err := rwf.Stat()
	fatalOnErr(t, errors.WithStack(err))

	if info.Name() != testFileName {
		t.Fatalf("Expected file name to be: '%s' Stat Found: %s", "testReadWriteFile", info.Name())
	}
	if info.IsDir() {
		t.Fatalf("Expected %s to be a file but stat said dir ", info.Name())
	}
	if info.Size() != 22 {
		t.Fatalf("Expected file %s to have size 22", sub)
	}

	// file.Sync
	fatalOnErr(t, errors.WithStack(rwf.Sync()))

	// file.Truncate
	fatalOnErr(t, errors.WithStack(rwf.Truncate(0)))
	info, err = rwf.Stat()
	fatalOnErr(t, errors.WithStack(err))
	if info.Size() != 0 {
		t.Fatalf("Expected truncated file %s to have size 0", sub)
	}
}
