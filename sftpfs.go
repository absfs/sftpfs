// Package sftpfs implements an absfs.Filer for SFTP (SSH File Transfer Protocol).
// It provides secure file operations over SSH using the github.com/pkg/sftp library.
package sftpfs

import (
	"io"
	iofs "io/fs"
	"os"
	"path/filepath"
	"time"

	"github.com/absfs/absfs"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// FileSystem implements absfs.Filer for SFTP protocol.
type FileSystem struct {
	client    sftpClientInterface
	sshClient sshClientInterface
}

// Config contains the configuration for connecting to an SFTP server.
type Config struct {
	Host     string        // Host address (e.g., "example.com:22")
	User     string        // Username for authentication
	Password string        // Password for authentication (if using password auth)
	Key      []byte        // Private key for authentication (if using key auth)
	Timeout  time.Duration // Connection timeout
}

// New creates a new SFTP filesystem with the given configuration.
func New(config *Config) (*FileSystem, error) {
	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Build SSH client config
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Timeout:         config.Timeout,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // WARNING: This skips host key verification
	}

	// Add authentication method
	if len(config.Key) > 0 {
		// Use key-based authentication
		signer, err := ssh.ParsePrivateKey(config.Key)
		if err != nil {
			return nil, err
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		// Use password authentication
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(config.Password)}
	}

	// Connect to SSH server
	sshClient, err := ssh.Dial("tcp", config.Host, sshConfig)
	if err != nil {
		return nil, err
	}

	// Create SFTP client
	client, err := sftp.NewClient(sshClient)
	if err != nil {
		sshClient.Close()
		return nil, err
	}

	return &FileSystem{
		client:    &sftpClientWrapper{client: client},
		sshClient: sshClient,
	}, nil
}

// newWithClients creates a FileSystem with injected clients for testing.
func newWithClients(sftpClient sftpClientInterface, sshClient sshClientInterface) *FileSystem {
	return &FileSystem{
		client:    sftpClient,
		sshClient: sshClient,
	}
}

// Close closes the SFTP connection.
func (fs *FileSystem) Close() error {
	if fs.client != nil {
		fs.client.Close()
	}
	if fs.sshClient != nil {
		return fs.sshClient.Close()
	}
	return nil
}

// OpenFile opens a file on the SFTP server.
func (fs *FileSystem) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	file, err := fs.client.OpenFile(name, flag)
	if err != nil {
		return nil, err
	}
	return &File{file: file, name: name, client: fs.client}, nil
}

// Mkdir creates a directory on the SFTP server.
func (fs *FileSystem) Mkdir(name string, perm os.FileMode) error {
	return fs.client.Mkdir(name)
}

// Remove removes a file or empty directory from the SFTP server.
func (fs *FileSystem) Remove(name string) error {
	return fs.client.Remove(name)
}

// Rename renames a file on the SFTP server.
func (fs *FileSystem) Rename(oldpath, newpath string) error {
	return fs.client.Rename(oldpath, newpath)
}

// Stat returns file info for a file on the SFTP server.
func (fs *FileSystem) Stat(name string) (os.FileInfo, error) {
	return fs.client.Stat(name)
}

// Chmod changes the mode of a file on the SFTP server.
func (fs *FileSystem) Chmod(name string, mode os.FileMode) error {
	return fs.client.Chmod(name, mode)
}

// Chtimes changes the access and modification times of a file on the SFTP server.
func (fs *FileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return fs.client.Chtimes(name, atime, mtime)
}

// Chown changes the owner and group of a file on the SFTP server.
func (fs *FileSystem) Chown(name string, uid, gid int) error {
	return fs.client.Chown(name, uid, gid)
}

// ReadDir reads the directory named by name and returns a list of directory entries.
func (fs *FileSystem) ReadDir(name string) (entries []iofs.DirEntry, err error) {
	infos, err := fs.client.ReadDir(name)
	if err != nil {
		return nil, err
	}

	entries = make([]iofs.DirEntry, len(infos))
	for i, info := range infos {
		entries[i] = &dirEntry{info: info}
	}
	return entries, nil
}

// ReadFile reads the file named by name and returns the contents.
func (fs *FileSystem) ReadFile(name string) ([]byte, error) {
	f, err := fs.OpenFile(name, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return io.ReadAll(f)
}

// Sub returns an fs.FS corresponding to the subtree rooted at dir.
func (fs *FileSystem) Sub(dir string) (iofs.FS, error) {
	return absfs.FilerToFS(fs, dir)
}

// ErrNotDir is returned when a path is expected to be a directory but is not.
var ErrNotDir = os.ErrInvalid

// Dial creates a new SFTP filesystem by dialing the specified host.
// This is a convenience function for simple password-based authentication.
func Dial(host, user, password string) (*FileSystem, error) {
	return New(&Config{
		Host:     host,
		User:     user,
		Password: password,
	})
}

// DialWithKey creates a new SFTP filesystem using SSH key authentication.
func DialWithKey(host, user string, privateKey []byte) (*FileSystem, error) {
	return New(&Config{
		Host: host,
		User: user,
		Key:  privateKey,
	})
}

// dirEntry implements fs.DirEntry for SFTP file info.
type dirEntry struct {
	info os.FileInfo
}

func (d *dirEntry) Name() string {
	return d.info.Name()
}

func (d *dirEntry) IsDir() bool {
	return d.info.IsDir()
}

func (d *dirEntry) Type() iofs.FileMode {
	return d.info.Mode().Type()
}

func (d *dirEntry) Info() (iofs.FileInfo, error) {
	return d.info, nil
}

// subFS implements a sub-filesystem rooted at a specific directory.
type subFS struct {
	parent *FileSystem
	root   string
}

func (s *subFS) joinPath(name string) string {
	return filepath.Join(s.root, name)
}

func (s *subFS) OpenFile(name string, flag int, perm os.FileMode) (absfs.File, error) {
	return s.parent.OpenFile(s.joinPath(name), flag, perm)
}

func (s *subFS) Mkdir(name string, perm os.FileMode) error {
	return s.parent.Mkdir(s.joinPath(name), perm)
}

func (s *subFS) Remove(name string) error {
	return s.parent.Remove(s.joinPath(name))
}

func (s *subFS) Rename(oldpath, newpath string) error {
	return s.parent.Rename(s.joinPath(oldpath), s.joinPath(newpath))
}

func (s *subFS) Stat(name string) (os.FileInfo, error) {
	return s.parent.Stat(s.joinPath(name))
}

func (s *subFS) Chmod(name string, mode os.FileMode) error {
	return s.parent.Chmod(s.joinPath(name), mode)
}

func (s *subFS) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return s.parent.Chtimes(s.joinPath(name), atime, mtime)
}

func (s *subFS) Chown(name string, uid, gid int) error {
	return s.parent.Chown(s.joinPath(name), uid, gid)
}

func (s *subFS) ReadDir(name string) ([]iofs.DirEntry, error) {
	return s.parent.ReadDir(s.joinPath(name))
}

func (s *subFS) ReadFile(name string) ([]byte, error) {
	return s.parent.ReadFile(s.joinPath(name))
}

func (s *subFS) Sub(dir string) (iofs.FS, error) {
	return absfs.FilerToFS(s, dir)
}
