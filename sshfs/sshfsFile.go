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

// ReadAt reads reads up to len(b) bytes at offset n
func (f *file) ReadAt(b []byte, n int64) (int, error) {
	n, err := f.Seek(n, io.SeekStart)
	if err != nil {
		return int(n), err
	}

	// rewind the file - we don't have access to underlying readAt
	defer f.Seek(0, io.SeekStart)
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
	if count < 1 {
		return nil, errors.New("Readdir count should be greater than 0. Path: " + f.Name())
	}
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
	infosLen := len(infos)
	outLen := 0
	switch {
	case infosLen == 0, infosLen == count:
		return infos, err
	case infosLen > count:
		outLen = count
	case infosLen < count:
		outLen = infosLen
	}
	return infos[:outLen], err
}

func (f *file) Readdirnames(count int) ([]string, error) {
	infos, err := f.Readdir(count)
	if err != nil {
		return nil, err
	}
	infosLen := len(infos)
	outLen := 0
	switch {
	case infosLen > count:
		outLen = count
	case infosLen < count:
		outLen = infosLen
	default:
		outLen = count
	}
	names := make([]string, outLen)
	for i := 0; i < outLen; i++ {
		names[i] = infos[i].Name()
	}

	return names, nil
}

func (f *file) Sync() error {
	return nil
}

func (f *file) WriteString(s string) (int, error) {
	return f.File.Write([]byte(s))
}
