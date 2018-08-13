package sshfs

import (
	"io"
	"os"

	"github.com/pkg/sftp"

	"github.com/spf13/afero"
)

type file struct {
	*sftp.File
}

func newSSHFile(f *sftp.File) afero.File {
	return &file{f}
}

func (f *file) ReadAt(b []byte, n int64) (int, error) {
	n, err := f.Seek(n, io.SeekStart)
	if err != nil {
		return int(n), err
	}

	return f.Read(b)
}

func (f *file) WriteAt(b []byte, n int64) (int, error) {
	return 0, nil
}

func (f *file) Readdir(count int) ([]os.FileInfo, error) {
	return []os.FileInfo{}, nil
}

func (f *file) Readdirnames(n int) ([]string, error) {
	return []string{}, nil
}

func (f *file) Sync() error {
	return nil
}

func (f *file) WriteString(s string) (int, error) {
	return f.File.Write([]byte(s))
}
