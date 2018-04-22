package sshfs_test

import (
	"log"
	"testing"

	"net/http"
	_ "net/http/pprof"

	"github.com/spf13/afero/sshfs"
)

var testPort = 22
var testSrv *testSSHServ

func ensureTestServer(t *testing.T) {
	if testSrv == nil {
		testSrv = NewTestSSHServ()
		log.Println("Running test server 1")
		testSrv.Listen(testPort)
	}
}

func TestMkdir(t *testing.T) {

	go http.ListenAndServe("localhost:8900", nil)
	// go ensureTestServer(t)
	log.Println("Running test server")
	fs, err := sshfs.New("10.1.3.244", testPort, "/tmp")
	if err != nil {
		t.Fatal("Failed to open a new ssh filesystem", err)
	}
	log.Println("Sending mkdir")

	err = fs.Mkdir("testdir1", 0744)
	if err != nil {
		t.Error("Failed mkdir", err)
	}
	log.Println("Done mkdir")
	// cmd := strings.Join(testSrv.commands, "::")
	// expected := "mkdir testdir1"
	// if cmd != expected {
	// 	t.Errorf("Mkdir command mismatched\n Expected: %s\nRan: %s", expected, cmd)
	// }
}
