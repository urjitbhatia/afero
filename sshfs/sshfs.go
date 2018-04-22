package sshfs

import (
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
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
func New(host string, port int, username, password string, root string) (afero.Fs, error) {
	// Todo : handle reconnecting broken connections...
	conn, err := connect(username, password, host, port)
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
	// The reason to use `install` is to be able to create a dir
	// and apply the right permissions atomically - if for some reason
	// We could have called a mkdir followed by a chmod call but if we crash right
	// after mkdir, we potentially leave an exposed directory on somebody's production
	// server.

	// TODO: check `install` is available on what platforms by default
	cmd := fmt.Sprintf("install -d -m %o %s/%s", perm, fs.Root, name)
	log.Println("Sending remote command: ", cmd, len(cmd))
	r, err := sess.CombinedOutput(cmd)
	if err == nil {
		log.Println("Error performing mkdir: ", r, err)
	}
	return nil
}

// MkdirAll creates all the directories
func (fs *Fs) MkdirAll(path string, perm os.FileMode) error {
	sess := fs.sessions.Get().(*session)
	if sess.err != nil {
		return sess.err
	}
	defer func() {
		sess.Close()
		fs.sessions.Put(sess)
	}()
	parts := []string{}
	for _, part := range strings.Split(path, "/") {
		if part == "" {
			// takes care of "//dir" or "/dir//dir1" etc
			continue
		}
		parts = append(parts, part)
	}

	return fs.Mkdir(strings.Join(parts, "/"), perm)
}

// Open opens the named file for reading.
func (fs *Fs) Open(name string) (afero.File, error) { return nil, nil }

// OpenFile is the generalized open call; most users will use Open or Create instead.
// It opens the named file with specified flag (O_RDONLY etc.)
// and perm (before umask), if applicable.
// If successful, methods on the returned File can be used for I/O.
// If there is an error, it will be of type *PathError.
func (fs *Fs) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	return nil, nil
}

// Remove removes the named file or directory. If there is an error, it will be of type *PathError.
func (fs *Fs) Remove(name string) error { return nil }

// RemoveAll removes path and any children it contains.
// It removes everything it can but returns the first error it encounters.
// If the path does not exist, RemoveAll returns nil (no error).
func (fs *Fs) RemoveAll(path string) error { return nil }

// Rename renames (moves) oldpath to newpath. If newpath already exists and is not a directory,
// Rename replaces it. OS-specific restrictions may apply when
// oldpath and newpath are in different directories.
// If there is an error, it will be of type *LinkError.
func (fs *Fs) Rename(oldname, newname string) error { return nil }

// Stat returns the FileInfo structure describing file. If there is an error, it will be of type *PathError.
func (fs *Fs) Stat(name string) (os.FileInfo, error) { return nil, nil }

// Name returns the name of the filesystem
func (fs *Fs) Name() string { return fs.String() }

// Chmod changes the mode of the named file to mode.
// If the file is a symbolic link, it changes the mode of the link's target.
// If there is an error, it will be of type *PathError.
// See docs in file.go?s=10557:10601#L326
func (fs *Fs) Chmod(name string, mode os.FileMode) error {
	sess := fs.sessions.Get().(*session)
	if sess.err != nil {
		return sess.err
	}
	defer func() {
		sess.Close()
		fs.sessions.Put(sess)
	}()
	cmd := fmt.Sprintf("chmod %o %s", mode, sanitizePath(fs.Root, name))
	r, err := sess.CombinedOutput(cmd)
	if err == nil {
		log.Println("Error performing chmod: ", r, err)
	}
	return nil
}

// Chtimes changes the access and modification times of the named
// file, similar to the Unix utime() or utimes() functions.
//
// The underlying filesystem may truncate or round the values to a
// less precise time unit.
// If there is an error, it will be of type *PathError.
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
