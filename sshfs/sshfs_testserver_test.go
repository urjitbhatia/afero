package sshfs_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/spf13/afero"

	"golang.org/x/crypto/ssh"
)

// testSSHServ provides an afero in-memory FS backed "server"
type testSSHServ struct {
	fs       afero.Fs
	commands []string
}

// NewTestSSHServ creates a new test ssh server backed by afero in-memory FS
func NewTestSSHServ() *testSSHServ {
	return &testSSHServ{afero.NewMemMapFs(), []string{}}
}

// Listen starts a test SSH server listens on given port.
func (ts *testSSHServ) Listen(port int) {
	config := &ssh.ServerConfig{
		// Define a function to run when a client attempts a password login
		// PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		// 	// Should use constant-time compare (or better, salt+hash) in a production setting.
		// 	if c.User() == "foo" && string(pass) == "bar" {
		// 		return nil, nil
		// 	}
		// 	return nil, fmt.Errorf("password rejected for %q", c.User())
		// },
		NoClientAuth: true,
		// You may also explicitly allow anonymous client authentication, though anon bash
		// sessions may not be a wise idea
		// NoClientAuth: true,
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
		// Discard all global out-of-band Requests
		go handleRequests(reqs)
		// Accept all channels
		for newChannel := range chans {
			// Technically, this should be a go routine but we will only test one thing at a time
			// so its ok
			log.Println("got a chan...")
			go ts.handleChannel(newChannel)
		}
	}
}

func handleRequests(reqs <-chan *ssh.Request) {
	log.Println("handling requests")
	for r := range reqs {
		log.Println("got payload from request: ", r.Payload)
	}
}

func (ts *testSSHServ) handleChannel(newChannel ssh.NewChannel) {
	log.Println("hankding ssh channel")
	// Only handle sessions here
	if t := newChannel.ChannelType(); t != "session" {
		newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
		return
	}
	// We will ignore requests
	connection, requests, err := newChannel.Accept()
	log.Println("Got a connection from chan ", connection)
	if err != nil {
		log.Printf("Could not accept channel (%s)", err)
		return
	}
	defer connection.Close()

	// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
	go func() {
		for req := range requests {
			switch req.Type {
			case "exec":
				// We only accept the default exec type requests
				log.Println("bad request ", string(req.Payload))
			default:
				log.Println("bad request ", req.Payload, req.Type)
			}
		}
	}()

	data, err := ioutil.ReadAll(connection)
	if err != nil {
		log.Println("Error reading data from connection", err)
		io.Copy(connection, bytes.NewReader([]byte("Error reading from connection")))
	}
	ts.commands = append(ts.commands, string(data))
	log.Println("Saving command: ", string(data))
}
