package sshfs_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/spf13/afero/sshfs"
)

var testPort = 0
var testFs afero.Fs

func getTestFs(t *testing.T) afero.Fs {
	if testFs == nil {
		sftp := testClientGoSvr(t)
		// memoize our test fs
		testFs = sshfs.NewWithClient("localhost", testPort, "", "", sftp)
	}
	return testFs
}

func fatalOnErr(t *testing.T, err error) {
	if err != nil {
		t.Fatal(err)
	}
}

func TestMkdir(t *testing.T) {
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, err)

	sub := path.Join(dir, "mkdir1")
	fatalOnErr(t, fs.Mkdir(sub, 0744))

	_, err = os.Lstat(sub)
	fatalOnErr(t, err)
}
func TestMkdirAll(t *testing.T) {
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, err)

	sub := path.Join(dir, "mkdirall1", "mkdirall2", "mkdirall3")
	fatalOnErr(t, fs.MkdirAll(sub, 0744))

	_, err = os.Lstat(sub)
	fatalOnErr(t, err)
}
func TestCreate(t *testing.T) {
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, err)

	sub := path.Join(dir, "testCreateFile")
	f, err := fs.Create(sub)
	fatalOnErr(t, err)

	defer f.Close()
	_, err = os.Lstat(sub)
	fatalOnErr(t, err)
}

func TestOpen(t *testing.T) {
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, err)

	sub := path.Join(dir, "testOpenFile")
	cf, err := fs.Create(sub)
	fatalOnErr(t, err)

	defer cf.Close()
	f, err := fs.Open(sub)
	fatalOnErr(t, err)

	defer f.Close()
	_, err = os.Lstat(sub)
	fatalOnErr(t, err)
}

func TestChmod(t *testing.T) {
	fs := getTestFs(t)

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, err)

	sub := path.Join(dir, "testChmodFile")
	cf, err := fs.Create(sub)
	fatalOnErr(t, err)
	defer cf.Close()

	// Set mode to 0666
	fatalOnErr(t, os.Chmod(sub, 0666))

	// Now use the sshfs to set it to 0644
	mod := os.FileMode(0644)
	err = fs.Chmod(sub, mod)
	fatalOnErr(t, err)

	info, err := os.Lstat(sub)
	fatalOnErr(t, err)

	if info.Mode() != mod {
		t.Fatalf("Expected file mode: %o, found: %o", mod, info.Mode())
	}
}
func TestChtimes(t *testing.T) {
	fs := getTestFs(t)

	testTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	dir, err := ioutil.TempDir("", "sftptest")
	fatalOnErr(t, err)

	sub := path.Join(dir, "testChtimesFile")
	cf, err := fs.Create(sub)
	fatalOnErr(t, err)
	defer cf.Close()

	//  change time
	err = fs.Chtimes(sub, testTime, testTime)
	fatalOnErr(t, err)

	info, err := os.Lstat(sub)
	fatalOnErr(t, err)

	if info.ModTime().Unix() != testTime.Unix() {
		t.Fatalf("Expected file mod time: %v, found: %v", testTime, info.ModTime())
	}
}
