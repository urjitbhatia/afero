package sshfs

import (
	"errors"
	"io"
	"os"

	"github.com/pkg/sftp"

	"github.com/spf13/afero"
)

type file struct {
	*sftp.File
	c *sftp.Client
}

func newSSHFile(f *sftp.File, c *sftp.Client) afero.File {
	return &file{f, c}
}

func (f *file) ReadAt(b []byte, n int64) (int, error) {
	n, err := f.Seek(n, io.SeekStart)
	if err != nil {
		return int(n), err
	}

	return f.Read(b)
}

func (f *file) WriteAt(b []byte, n int64) (int, error) {
	n, err := f.Seek(n, io.SeekStart)
	if err != nil {
		return int(n), err
	}
	return f.Write(b)
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &os.PathError{Err: errors.New("Non dir file"), Op: "readdir", Path: f.Name()}
	}
	return f.c.ReadDir(f.Name())
}

func (f *file) Readdirnames(n int) ([]string, error) {
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, &os.PathError{Err: errors.New("Non dir file"), Op: "readdir", Path: f.Name()}
	}
	infos, err := f.c.ReadDir(f.Name())
	if err != nil {
		return nil, err
	}
	// TODO - paginate on n
	names := make([]string, len(infos))
	for i, n := range infos {
		names[i] = n.Name()
	}

	return names, nil
}

func (f *file) Sync() error {
	return nil
}

func (f *file) WriteString(s string) (int, error) {
	return f.File.Write([]byte(s))
}
