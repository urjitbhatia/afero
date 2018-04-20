package sshfs

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
)

type SSHFS struct {
	Host     string
	Port     int
	Root     string
	conn     *ssh.Client
	config   *ssh.Config
	sessions sync.Pool
}

func NewSSHFS(host string, port int, root string) (afero.Fs, error) {
	conn, err := connect("", "", host, port)
	if err != nil {
		return nil, err
	}
	sessionPool := sync.Pool{
		New: func() interface{} {
			s, err := conn.NewSession()
			if err != nil {
				return s
			}
			return nil
		},
	}
	return &SSHFS{Host: host,
		Port:     port,
		Root:     root,
		sessions: sessionPool}, nil
}

// Stringer
func (fs *SSHFS) String() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return fmt.Sprintf("SSH_FS@%s:%s::%s@@%s", fs.Host, fs.Port, fs.Root, hostname)
}

// Create a new file
func (fs *SSHFS) Create(name string) (afero.File, error) { return nil, nil }

// Mkdir creates a new dir
func (fs *SSHFS) Mkdir(name string, perm os.FileMode) error {
	sess := fs.sessions.Get().(*ssh.Session)
	defer fs.sessions.Put(sess)
	return sess.Run(fmt.Sprintf("mkdir %s/%s", fs.Root, name))
}

// MkdirAll creates all the directories
func (fs *SSHFS) MkdirAll(path string, perm os.FileMode) error {
	sess := fs.sessions.Get().(*ssh.Session)
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

func (fs *SSHFS) Open(name string) (afero.File, error) { return nil, nil }
func (fs *SSHFS) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return nil, nil
}
func (fs *SSHFS) Remove(name string) error              { return nil }
func (fs *SSHFS) RemoveAll(path string) error           { return nil }
func (fs *SSHFS) Rename(oldname, newname string) error  { return nil }
func (fs *SSHFS) Stat(name string) (os.FileInfo, error) { return nil, nil }
func (fs *SSHFS) Name() string                          { return fs.String() }
func (fs *SSHFS) Chmod(name string, mode os.FileMode) error {
	return nil
}
func (fs *SSHFS) Chtimes(name string, atime time.Time, mtime time.Time) error { return nil }

func connect(user, password, host string, port int) (*ssh.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		err          error
	)

	// private key
	privkey := ssh.ParsePrivateKey

	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.FixedHostKey())

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
	}

	// connet to ssh
	addr = fmt.Sprintf("%s:%d", host, port)

	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	return sshClient, nil
}
