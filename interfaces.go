package sftpfs

import (
	"os"
	"time"
)

// sftpClientInterface defines the methods we use from *sftp.Client.
// This enables mocking for unit tests.
type sftpClientInterface interface {
	Close() error
	OpenFile(path string, f int) (sftpFileInterface, error)
	Mkdir(path string) error
	Remove(path string) error
	Rename(oldpath, newpath string) error
	Stat(path string) (os.FileInfo, error)
	Chmod(path string, mode os.FileMode) error
	Chtimes(path string, atime, mtime time.Time) error
	Chown(path string, uid, gid int) error
	ReadDir(path string) ([]os.FileInfo, error)
}

// sftpFileInterface defines the methods we use from *sftp.File.
type sftpFileInterface interface {
	Read(b []byte) (int, error)
	ReadAt(b []byte, off int64) (int, error)
	Write(b []byte) (int, error)
	WriteAt(b []byte, off int64) (int, error)
	Seek(offset int64, whence int) (int64, error)
	Close() error
	Stat() (os.FileInfo, error)
	Truncate(size int64) error
}

// sshClientInterface defines the methods we use from *ssh.Client.
type sshClientInterface interface {
	Close() error
}
