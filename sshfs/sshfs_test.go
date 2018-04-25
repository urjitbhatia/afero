package sshfs_test

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"github.com/spf13/afero/sshfs"
)

var testPort = 22000

func getTestFs(sftp *sftp.Client) afero.Fs {
	return sshfs.NewWithClient("localhost", testPort, "", "", sftp)
}

func TestMkdir(t *testing.T) {
	sftp := testClientSvr(t)
	defer sftp.Close()
	fs := getTestFs(sftp)

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "mkdir1")
	if err := fs.Mkdir(sub, 0744); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(sub); err != nil {
		t.Fatal(err)
	}
}
func TestMkdirAll(t *testing.T) {
	sftp := testClientSvr(t)
	defer sftp.Close()
	fs := getTestFs(sftp)

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "mkdirall1", "mkdirall2", "mkdirall3")
	if err := fs.MkdirAll(sub, 0744); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Lstat(sub); err != nil {
		t.Fatal(err)
	}
}
func TestCreate(t *testing.T) {
	sftp := testClientSvr(t)
	defer sftp.Close()
	fs := getTestFs(sftp)

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "testCreateFile")
	f, err := fs.Create(sub)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := os.Lstat(sub); err != nil {
		t.Fatal(err)
	}
}

func TestOpen(t *testing.T) {
	sftp := testClientSvr(t)
	defer sftp.Close()
	fs := getTestFs(sftp)

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "testOpenFile")
	cf, err := fs.Create(sub)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()
	f, err := fs.Open(sub)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := os.Lstat(sub); err != nil {
		t.Fatal(err)
	}
}

func TestChmod(t *testing.T) {
	sftp := testClientSvr(t)
	defer sftp.Close()
	fs := getTestFs(sftp)

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "testChmodFile")
	cf, err := fs.Create(sub)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()

	// Set mode to 0666
	err = os.Chmod(sub, 0666)
	if err != nil {
		t.Fatal(err)
	}

	// Now use the sshfs to set it to 0644
	mod := os.FileMode(0644)
	err = fs.Chmod(sub, mod)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(sub)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode() != mod {
		t.Fatalf("Expected file mode: %o, found: %o", mod, info.Mode())
	}
}
func TestChtimes(t *testing.T) {
	sftp := testClientSvr(t)
	defer sftp.Close()
	fs := getTestFs(sftp)

	testTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	dir, err := ioutil.TempDir("", "sftptest")
	if err != nil {
		t.Fatal(err)
	}
	sub := path.Join(dir, "testChtimesFile")
	cf, err := fs.Create(sub)
	if err != nil {
		t.Fatal(err)
	}
	defer cf.Close()

	//  change time
	err = fs.Chtimes(sub, testTime, testTime)
	if err != nil {
		t.Fatal(err)
	}

	info, err := os.Lstat(sub)
	if err != nil {
		t.Fatal(err)
	}
	if info.ModTime().Unix() != testTime.Unix() {
		t.Fatalf("Expected file mod time: %v, found: %v", testTime, info.ModTime())
	}
}
