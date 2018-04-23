package sshfs_test

import (
	"log"
	"testing"
	"time"

	"github.com/spf13/afero/sshfs"
)

var testPort = 22000
var testSrv *testSSHServ
var sshCmdCaptor = make(chan string)

func ensureTestServer(t *testing.T) {
	if testSrv == nil {
		testSrv = NewTestSSHServ(sshCmdCaptor)
		log.Println("Running test server")
		testSrv.Listen(testPort)
	}
}

func TestMkdir(t *testing.T) {
	go ensureTestServer(t)
	fs, err := sshfs.New("localhost", testPort, "", "", "/tmp")
	if err != nil {
		t.Fatal("Failed to open a new ssh filesystem", err)
	}

	err = fs.Mkdir("testdir", 0744)
	if err != nil {
		t.Error("Failed mkdir", err)
	}
	cmd := <-sshCmdCaptor
	expected := "install -d -m 744 /tmp/testdir"

	if cmd != expected {
		t.Errorf("Mkdir command mismatched\nExpected:\t%s\nGot:\t%s", expected, cmd)
	}
}
func TestMkdirAll(t *testing.T) {
	go ensureTestServer(t)
	fs, err := sshfs.New("localhost", testPort, "", "", "/tmp")
	if err != nil {
		t.Fatal("Failed to open a new ssh filesystem", err)
	}

	err = fs.MkdirAll("/testdir/subdir/subsubdir", 0744)
	if err != nil {
		t.Error("Failed mkdirall", err)
	}
	cmd := <-sshCmdCaptor
	expected := "install -d -m 744 /tmp/testdir/subdir/subsubdir"

	if cmd != expected {
		t.Errorf("Mkdir command mismatched\nExpected:\t%s\nGot:\t%s", expected, cmd)
	}
}
func TestChmod(t *testing.T) {
	go ensureTestServer(t)
	fs, err := sshfs.New("localhost", testPort, "", "", "/tmp")
	if err != nil {
		t.Fatal("Failed to open a new ssh filesystem", err)
	}

	err = fs.Chmod("/testdir//foo", 0744)
	if err != nil {
		t.Error("Failed chmod", err)
	}
	cmd := <-sshCmdCaptor
	expected := "chmod 744 /tmp/testdir/foo"

	if cmd != expected {
		t.Errorf("Mkdir command mismatched\nExpected:\t%s\nGot:\t%s", expected, cmd)
	}
}
func TestChtimes(t *testing.T) {
	go ensureTestServer(t)
	fs, err := sshfs.New("localhost", testPort, "", "", "/tmp")
	// fs, err := sshfs.New("10.1.3.244", 22, "ubuntu", "", "/tmp")
	if err != nil {
		t.Fatal("Failed to open a new ssh filesystem", err)
	}
	testTime, err := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	if err != nil {
		t.Error("Cannot parse test time ", err)
	}
	err = fs.Chtimes("/testdir1/foo", testTime, testTime)
	if err != nil {
		t.Error("Failed chtimes", err)
	}
	acmd := <-sshCmdCaptor
	aexpected := "touch -a -t 200601021504.05 /tmp/testdir1/foo"
	mcmd := <-sshCmdCaptor
	mexpected := "touch -m -t 200601021504.05 /tmp/testdir1/foo"

	if acmd != aexpected {
		t.Errorf("Mkdir command mismatched\nExpected:\t%s\nGot:\t%s", aexpected, acmd)
	}
	if mcmd != mexpected {
		t.Errorf("Mkdir command mismatched\nExpected:\t%s\nGot:\t%s", mexpected, mcmd)
	}
}
