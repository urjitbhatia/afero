package sshfs_test

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestReadWrite(t *testing.T) {
	testFileName := "testReadWriteFile"
	// Get test file system
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, err)
	sub := path.Join(dir, testFileName)

	// Create a test file
	rwf, err := fs.Create(sub)
	fatalOnErr(t, err)

	// Write something to it
	_, err = rwf.WriteString("Hello World")
	fatalOnErr(t, err)

	// Close and flush
	fatalOnErr(t, rwf.Close())

	// Open it and read what we wrote
	rwf, err = fs.Open(sub)
	fatalOnErr(t, err)
	defer rwf.Close()

	buf, err := ioutil.ReadAll(rwf)
	fatalOnErr(t, err)

	bufs := string(buf)
	if bufs != "Hello World" {
		t.Fatalf("Expected file to contain: 'Hello World' Found: %s", bufs)
	}

	buf = make([]byte, 5)       // Read World
	_, err = rwf.ReadAt(buf, 6) // Skip 'Hello '
	fatalOnErr(t, err)

	bufs = string(buf)
	if bufs != "World" {
		t.Fatalf("Expected file to contain: 'World' Found: %s", bufs)
	}

	// Read name
	if rwf.Name() != sub {
		t.Fatalf("Expected file name to be: '%s' Found: %s", sub, rwf.Name())
	}

	// Some Stat file checks
	info, err := rwf.Stat()
	fatalOnErr(t, err)

	if info.Name() != testFileName {
		t.Fatalf("Expected file name to be: '%s' Stat Found: %s", "testReadWriteFile", info.Name())
	}
	if info.IsDir() {
		t.Fatalf("Expected %s to be a file but stat said dir ", info.Name())
	}
	if info.Size() != 11 {
		t.Fatalf("Expected file %s to have size 11", sub)
	}
}
