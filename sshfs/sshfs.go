package sshfs

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/pkg/sftp"
	"github.com/spf13/afero"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
)

// Fs is an afero filesystem over ssh
type Fs struct {
	Host   string
	Port   int
	client *sftp.Client
}

// New provides an afero filesystem over ssh
func New(host string, port int, username, password string) (afero.Fs, error) {
	// Todo : handle reconnecting broken connections...
	conn, err := connect(username, password, host, port)
	if err != nil {
		return nil, err
	}
	client, err := sftp.NewClient(conn)
	if err != nil {
		return nil, err
	}
	return &Fs{Host: host,
		Port:   port,
		client: client}, nil
}

func NewWithClient(host string, port int, username, password string, client *sftp.Client) afero.Fs {
	return &Fs{Host: host,
		Port:   port,
		client: client}
}

// String representation of this fs
func (fs *Fs) String() string {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "localhost"
	}
	return fmt.Sprintf("SSH_FS@%s:%d::%s@@%s", fs.Host, fs.Port, hostname)
}

// Create a new file
func (fs *Fs) Create(name string) (afero.File, error) {
	f, err := fs.client.Create(name)
	if err != nil {
		return nil, err
	}
	return newSSHFile(f), nil
}

// Mkdir creates a new dir
func (fs *Fs) Mkdir(name string, perm os.FileMode) error {
	err := fs.client.Mkdir(name)
	if err != nil {
		return err
	}
	return fs.client.Chmod(name, perm)
}

// MkdirAll creates all the directories
func (fs *Fs) MkdirAll(path string, perm os.FileMode) error {
	parts := ""
	for _, p := range strings.Split(path, "/") {
		if p == "" {
			continue
		}
		parts += "/" + p
		dir, err := fs.client.Stat(parts)
		if err == nil {
			if !dir.IsDir() {
				return fmt.Errorf("Found a non-directory file on path: %s", parts)
			}
			continue
		}
		err = fs.Mkdir(parts, perm)
		if err != nil {
			return err
		}
	}
	return nil
}

// Open opens the named file for reading.
func (fs *Fs) Open(name string) (afero.File, error) {
	f, err := fs.client.Open(name)
	if err != nil {
		return nil, err
	}
	af := newSSHFile(f)
	return af, err
}

// OpenFile is the generalized open call; most users will use Open or Create instead.
// It opens the named file with specified flag (O_RDONLY etc.)
// and perm (before umask), if applicable.
// If successful, methods on the returned File can be used for I/O.
// If there is an error, it will be of type *PathError.
func (fs *Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return nil, nil
}

// Remove removes the named file or directory. If there is an error, it will be of type *PathError.
func (fs *Fs) Remove(name string) error { return fs.client.Remove(name) }

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error it encounters.
// If the path does not exist, RemoveAll returns nil (no error).
func (fs *Fs) RemoveAll(path string) error { return fs.client.RemoveDirectory(path) }

// Rename renames (moves) oldpath to newpath. If newpath already exists and is not a directory,
// Rename replaces it. OS-specific restrictions may apply when
// oldpath and newpath are in different directories.
// If there is an error, it will be of type *LinkError.
func (fs *Fs) Rename(oldname, newname string) error { return fs.client.Rename(oldname, newname) }

// Stat returns the FileInfo structure describing file. If there is an error, it will be of type *PathError.
func (fs *Fs) Stat(name string) (os.FileInfo, error) { return fs.client.Stat(name) }

// Name returns the name of the filesystem
func (fs *Fs) Name() string { return fs.String() }

// Chmod changes the mode of the named file to mode.
func (fs *Fs) Chmod(name string, mode os.FileMode) error { return fs.client.Chmod(name, mode) }

// Chtimes changes the access and modification times of the named
// file, similar to the Unix utime() or utimes() functions.
func (fs *Fs) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return fs.client.Chtimes(name, atime, mtime)
}

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
			sshAgent(),
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

// sshAgent creates an auth method using the ssh agent sock if available
func sshAgent() ssh.AuthMethod {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err == nil {
		return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
	}
	log.Println("Failed to instantiate ssh from SSH_AUTH_SOCK", err)
	return nil
}

func sanitizePath(root, path string) string {
	parts := append([]string{}, root)
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			// takes care of "//dir" or "/dir//dir1" etc
			continue
		}
		parts = append(parts, part)
	}

	return strings.Join(parts, "/")
}
