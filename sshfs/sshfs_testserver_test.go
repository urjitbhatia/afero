package sshfs_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"

	"github.com/spf13/afero"

	"golang.org/x/crypto/ssh"
)

// testSSHServ provides an afero in-memory FS backed "server"
type testSSHServ struct {
	fs        afero.Fs
	cmdCaptor chan string
}

// NewTestSSHServ creates a new test ssh server backed by afero in-memory FS
func NewTestSSHServ(cmdCaptor chan string) *testSSHServ {
	return &testSSHServ{afero.NewMemMapFs(), cmdCaptor}
}

// Listen starts a test SSH server listens on given port.
func (ts *testSSHServ) Listen(port int) {
	config := &ssh.ServerConfig{
		// This is a *TEST* ssh server without auth, implement a PasswordCallback etc if you want
		// to copy this code for a production system
		NoClientAuth: true,
	}

	// You can generate a keypair with 'ssh-keygen -t rsa'
	privateBytes, err := ioutil.ReadFile("./test_id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key (./test_id_rsa)")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be accepted.
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		log.Fatal("Failed to listen for a connection", err)
	}

	// Accept all connections
	log.Print("Listening on ", listener.Addr().String())
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		log.Println("Got a connection from", tcpConn.RemoteAddr().String())
		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// The incoming Request channel must be serviced.
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		for newChannel := range chans {
			// Technically, this should be a go routine but we will only test one thing at a time
			// so its ok
			go ts.handleChannel(newChannel)
		}
	}
}

func (ts *testSSHServ) handleChannel(newChannel ssh.NewChannel) {
	// Only handle sessions here
	if t := newChannel.ChannelType(); t != "session" {
		log.Println("Rejecting unknown channel type", t)
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}
	defer channel.Close()

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	// We only handle "exec" requests for this test setup

	// This should be a go routine to multiplex on channels but we don't need that
	// for testing
	for req := range requests {
		log.Printf("Handling a request of Type: %s Payload: %s", req.Type, req.Payload)
		switch req.Type {
		case "exec":
			// We only accept the default exec type requests for this test
			go ts.recordCommand(req.Payload)
			req.Reply(true, []byte("ok"))
			channel.Close()
			channel.CloseWrite()
		default:
			req.Reply(false, []byte("expected request type: exec"))
		}
	}
}

// Capture the prev request we saw and post it on this channel - for testing
func (ts *testSSHServ) recordCommand(cmd []byte) {
	// Somehow the wire protocol is prefixing 4 empty bytes at the beginning of the payload :(
	ts.cmdCaptor <- string(cmd[4:])
}
