package sshfs_test

import (
	"log"
	"testing"

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

	err = fs.Mkdir("testdir1", 0744)
	if err != nil {
		t.Error("Failed mkdir", err)
	}
	cmd := <-sshCmdCaptor
	expected := "install -d -m 744 /tmp/testdir1"

	if cmd != expected {
		t.Errorf("Mkdir command mismatched\nExpected:\t%s\nGot:\t%s", expected, cmd)
	}
}
