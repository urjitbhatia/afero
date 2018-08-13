package sshfs_test

import (
	"io/ioutil"
	"path"
	"testing"
)

func TestReadWrite(t *testing.T) {
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "testReadWriteFile")
	// Create a test file
	rwf, err := fs.Create(sub)
	if err != nil {
		t.Fatal(err)
	}
	// Write something to it
	_, err = rwf.WriteString("Hello World")
	if err != nil {
		t.Fatal(err)
	}
	// Close and flush
	err = rwf.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Open it and read what we wrote
	rwf, err = fs.Open(sub)
	if err != nil {
		t.Fatal(err)
	}
	defer rwf.Close()
	buf, err := ioutil.ReadAll(rwf)
	if err != nil {
		t.Fatal(err)
	}
	bufs := string(buf)
	if bufs != "Hello World" {
		t.Fatalf("Expected file to contain: 'Hello World' Found: %s", bufs)
	}
}
func TestReadOps(t *testing.T) {
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "testReadWriteFile")
	// Create a test file
	rwf, err := fs.Create(sub)
	if err != nil {
		t.Fatal(err)
	}
	// Write something to it
	_, err = rwf.WriteString("Hello World")
	if err != nil {
		t.Fatal(err)
	}
	// Close and flush
	err = rwf.Close()
	if err != nil {
		t.Fatal(err)
	}

	// Open it and read what we wrote
	rwf, err = fs.Open(sub)
	if err != nil {
		t.Fatal(err)
	}
	defer rwf.Close()
	buf := make([]byte, 5)      // Read World
	_, err = rwf.ReadAt(buf, 6) // Skip 'Hello '
	if err != nil {
		t.Fatal(err)
	}
	bufs := string(buf)
	if bufs != "World" {
		t.Fatalf("Expected file to contain: 'World' Found: %s", bufs)
	}
}
