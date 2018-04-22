package sshfs

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Fs is an afero filesystem over ssh
type Fs struct {
	Host     string
	Port     int
	Root     string
	conn     *ssh.Client
	config   *ssh.Config
	sessions *sync.Pool
}

type session struct {
	err error
	*ssh.Session
}

// New provides an afero filesystem over ssh
func New(host string, port int, root string) (afero.Fs, error) {
	conn, err := connect("ubuntu", "", host, port)
	if err != nil {
		return nil, err
	}
	sessionPool := sync.Pool{
		New: func() interface{} {
			s, err := conn.NewSession()
			return &session{err, s}
		},
	}
	return &Fs{Host: host,
		Port:     port,
		Root:     root,
		conn:     conn,
		sessions: &sessionPool}, nil
}

// String representation of this fs
func (fs *Fs) String() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return fmt.Sprintf("SSH_FS@%s:%d::%s@@%s", fs.Host, fs.Port, fs.Root, hostname)
}

// Create a new file
func (fs *Fs) Create(name string) (afero.File, error) { return nil, nil }

// Mkdir creates a new dir
func (fs *Fs) Mkdir(name string, perm os.FileMode) error {
	sess := fs.sessions.Get().(*session)
	if sess.err != nil {
		return sess.err
	}
	defer func() {
		sess.Close()
		fs.sessions.Put(sess)
	}()
	cmd := fmt.Sprintf("install -d -m %o %s/%s", perm, fs.Root, name)
	log.Println("Sending remote command: ", cmd)
	r, err := sess.CombinedOutput(cmd)
	if err == nil {
		log.Println("mkdir got: ", r)
	}
	return nil
}

// MkdirAll creates all the directories
func (fs *Fs) MkdirAll(path string, perm os.FileMode) error {
	sess := fs.sessions.Get().(*session)
	if sess.err != nil {
		return sess.err
	}
	defer fs.sessions.Put(sess)

	pathSoFar := fs.Root
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			// takes care of "//dir" or "/dir//dir1" etc
			continue
		}
		pathSoFar = pathSoFar + "/" + part
		if err := fs.Mkdir(pathSoFar, perm); err != nil {
			return err
		}
	}
	return nil
}

func (fs *Fs) Open(name string) (afero.File, error) { return nil, nil }
func (fs *Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return nil, nil
}
func (fs *Fs) Remove(name string) error              { return nil }
func (fs *Fs) RemoveAll(path string) error           { return nil }
func (fs *Fs) Rename(oldname, newname string) error  { return nil }
func (fs *Fs) Stat(name string) (os.FileInfo, error) { return nil, nil }
func (fs *Fs) Name() string                          { return fs.String() }
func (fs *Fs) Chmod(name string, mode os.FileMode) error {
	return nil
}
func (fs *Fs) Chtimes(name string, atime time.Time, mtime time.Time) error { return nil }

func connect(user, password, host string, port int) (*ssh.Client, error) {
	var (
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)

	clientConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			SSHAgent(),
		},
		Timeout:         100 * time.Millisecond,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// connect to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}
	log.Println("Got an ssh client", sshClient.RemoteAddr().String())
	return sshClient, nil
}

func SSHAgent() ssh.AuthMethod {
	if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	} else {
		log.Println("Failed to instantiate ssh from SSH_AUTH_SOCK", err)
	}
	return nil
}

// syscallMode returns the syscall-specific mode bits from Go's portable mode bits.
// From: https://golang.org/src/os/file_posix.go
func syscallMode(i os.FileMode) (o uint32) {
	o |= uint32(i.Perm())
	if i&os.ModeSetuid != 0 {
		o |= syscall.S_ISUID
	}
	if i&os.ModeSetgid != 0 {
		o |= syscall.S_ISGID
	}
	if i&os.ModeSticky != 0 {
		o |= syscall.S_ISVTX
	}
	// No mapping for Go's ModeTemporary (plan9 only).
	return
}
