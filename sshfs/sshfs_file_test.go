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

	// file.ReadAt - also that readAt leaves fd pointing back to start
	buf = make([]byte, 5)       // Read World
	_, err = rwf.ReadAt(buf, 6) // Skip 'Hello '
	fatalOnErr(t, errors.WithStack(err))

	bufs = string(buf)
	if bufs != "World" {
		t.Fatalf("Expected file to contain: 'World' Found: %s", bufs)
	}
	// follow ReadAt with a ReadAll to test file is back to start after readAt
	buf, err = ioutil.ReadAll(rwf)
	fatalOnErr(t, errors.WithStack(err))

	bufs = string(buf)
	if bufs != "Hello World golang.org" {
		t.Fatalf("Expected file to contain: 'Hello World golang.org' Found: %s", bufs)
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

	// file.ReadDir - should err for non-dir file
	_, err = rwf.Readdir(2)
	e, _ := err.(*os.PathError)
	if e.Err.Error() != "Non dir file" {
		t.Fatal("Expected os.Path performing Readdir on non-dir file", err)
	}

	// file.Readdirnames - should err for non-dir file
	_, err = rwf.Readdirnames(2)
	e, _ = err.(*os.PathError)
	if e.Err.Error() != "Non dir file" {
		t.Fatal("Expected os.Path performing Readdir on non-dir file", err)
	}
}

func TestDirCalls(t *testing.T) {
	testDirName := "testReadWriteDir"
	// Get test file system
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, errors.WithStack(err))
	sub := path.Join(dir, testDirName)

	// Create dir
	fatalOnErr(t, errors.WithStack((fs.Mkdir(sub, os.ModeDir|os.FileMode(0770)))))

	// Get dir handle
	d, err := fs.Open(sub)
	fatalOnErr(t, errors.WithStack(err))

	// Readdir
	info, err := d.Readdir(10)
	fatalOnErr(t, errors.WithStack(err))
	if len(info) != 0 {
		t.Fatalf("Expected empty dir to return 0 info items on readdir. Got %+v", info)
	}

	// Readdir with count <= 0 should err
	_, err = d.Readdir(-1)
	if err == nil {
		t.Fatalf("Expected readdir with count 0 to err")
	}

	// Now create files in the dir
	_, err = fs.Create(path.Join(sub, "testFile1"))
	fatalOnErr(t, errors.WithStack(err))
	_, err = fs.Create(path.Join(sub, "testFile2"))
	fatalOnErr(t, errors.WithStack(err))

	// Now Readdir should return info about the file we created (count > num files)
	info, err = d.Readdir(10)
	fatalOnErr(t, errors.WithStack(err))
	if len(info) != 2 {
		t.Fatalf("Expected non-empty dir to return 2 info items on readdir")
	}
	if info[0].Name() != "testFile1" {
		t.Fatalf("Expected readdir to list testFile")
	}

	// Now Readdir should return info about the file we created (count < num files)
	info, err = d.Readdir(1)
	fatalOnErr(t, errors.WithStack(err))
	if len(info) != 1 {
		t.Fatalf("Expected non-empty dir to return 1 (count 1) info items on readdir")
	}
	if info[0].Name() != "testFile1" {
		t.Fatalf("Expected readdir to list testFile")
	}

	// Readdirnames - count > num items
	names, err := d.Readdirnames(10)
	fatalOnErr(t, errors.WithStack(err))
	if len(names) != 2 {
		t.Fatalf("Expected non-empty dir to return 2 (count 2) name items on readdirnames")
	}
	// Readdirnames - count < num items
	names, err = d.Readdirnames(1)
	fatalOnErr(t, errors.WithStack(err))
	if len(names) != 1 {
		t.Fatalf("Expected non-empty dir to return 1 (count 1) name items on readdirnames")
	}
}
