package sftpfs

import (
	"os"
	"time"

	"github.com/pkg/sftp"
)

// sftpClientWrapper wraps *sftp.Client to implement sftpClientInterface.
type sftpClientWrapper struct {
	client *sftp.Client
}

func (w *sftpClientWrapper) Close() error {
	return w.client.Close()
}

func (w *sftpClientWrapper) OpenFile(path string, f int) (sftpFileInterface, error) {
	file, err := w.client.OpenFile(path, f)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (w *sftpClientWrapper) Mkdir(path string) error {
	return w.client.Mkdir(path)
}

func (w *sftpClientWrapper) Remove(path string) error {
	return w.client.Remove(path)
}

func (w *sftpClientWrapper) Rename(oldpath, newpath string) error {
	return w.client.Rename(oldpath, newpath)
}

func (w *sftpClientWrapper) Stat(path string) (os.FileInfo, error) {
	return w.client.Stat(path)
}

func (w *sftpClientWrapper) Chmod(path string, mode os.FileMode) error {
	return w.client.Chmod(path, mode)
}

func (w *sftpClientWrapper) Chtimes(path string, atime, mtime time.Time) error {
	return w.client.Chtimes(path, atime, mtime)
}

func (w *sftpClientWrapper) Chown(path string, uid, gid int) error {
	return w.client.Chown(path, uid, gid)
}

func (w *sftpClientWrapper) ReadDir(path string) ([]os.FileInfo, error) {
	return w.client.ReadDir(path)
}
