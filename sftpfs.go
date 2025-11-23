// Package sftpfs implements an absfs.Filer for SFTP (SSH File Transfer Protocol).
// It provides secure file operations over SSH using the github.com/pkg/sftp library.
package sftpfs

import (
	"fmt"
	"os"
	"time"

	"github.com/absfs/absfs"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// wrapError wraps an error with operation and path context
func wrapError(op, path string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("sftpfs.%s(%s): %w", op, path, err)
}

// wrapErrorf wraps an error with formatted context
func wrapErrorf(format string, args ...interface{}) error {
	return fmt.Errorf("sftpfs: "+format, args...)
}

// FileSystem implements absfs.FileSystem for SFTP protocol.
type FileSystem struct {
	client    *sftp.Client
	sshClient *ssh.Client
	cwd       string // current working directory
}

// Config contains the configuration for connecting to an SFTP server.
type Config struct {
	Host            string              // Host address (e.g., "example.com:22")
	User            string              // Username for authentication
	Password        string              // Password for authentication (if using password auth)
	Key             []byte              // Private key for authentication (if using key auth)
	Timeout         time.Duration       // Connection timeout
	HostKeyCallback ssh.HostKeyCallback // Host key verification callback (defaults to InsecureIgnoreHostKey if not set)
	MaxRetries      int                 // Maximum number of connection retry attempts (0 = no retries, default = 3)
	RetryDelay      time.Duration       // Initial delay between retries (default = 1 second, uses exponential backoff)
}

// New creates a new SFTP filesystem with the given configuration.
func New(config *Config) (*FileSystem, error) {
	// Set default timeout if not specified
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Set default retry parameters
	maxRetries := config.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3 // Default to 3 retries
	}

	retryDelay := config.RetryDelay
	if retryDelay == 0 {
		retryDelay = 1 * time.Second // Default to 1 second initial delay
	}

	// Set default host key callback if not specified
	hostKeyCallback := config.HostKeyCallback
	if hostKeyCallback == nil {
		// WARNING: This skips host key verification and is vulnerable to MITM attacks
		// For production use, provide a proper HostKeyCallback in the Config
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	// Build SSH client config
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Timeout:         config.Timeout,
		HostKeyCallback: hostKeyCallback,
	}

	// Add authentication method
	if len(config.Key) > 0 {
		// Use key-based authentication
		signer, err := ssh.ParsePrivateKey(config.Key)
		if err != nil {
			return nil, wrapErrorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else {
		// Use password authentication
		sshConfig.Auth = []ssh.AuthMethod{ssh.Password(config.Password)}
	}

	// Attempt connection with retry logic
	var sshClient *ssh.Client
	var client *sftp.Client
	var lastErr error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Connect to SSH server
		sshClient, lastErr = ssh.Dial("tcp", config.Host, sshConfig)
		if lastErr == nil {
			// Create SFTP client
			client, lastErr = sftp.NewClient(sshClient)
			if lastErr == nil {
				// Success!
				break
			}
			// SFTP client creation failed, close SSH client and retry
			sshClient.Close()
		}

		// If this was the last attempt, don't sleep
		if attempt < maxRetries {
			// Exponential backoff: delay * 2^attempt
			sleepDuration := retryDelay * time.Duration(1<<uint(attempt))
			time.Sleep(sleepDuration)
		}
	}

	if lastErr != nil {
		return nil, wrapErrorf("failed to connect after %d attempts: %w", maxRetries+1, lastErr)
	}

	return &FileSystem{
		client:    client,
		sshClient: sshClient,
		cwd:       "/", // Start at root directory
	}, nil
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
		return nil, wrapError("OpenFile", name, err)
	}

	// If creating a new file, set the permissions
	if flag&os.O_CREATE != 0 {
		if err := fs.client.Chmod(name, perm); err != nil {
			file.Close()
			return nil, wrapError("OpenFile.Chmod", name, err)
		}
	}

	return &File{file: file, name: name, client: fs.client}, nil
}

// Mkdir creates a directory on the SFTP server.
func (fs *FileSystem) Mkdir(name string, perm os.FileMode) error {
	if err := fs.client.Mkdir(name); err != nil {
		return wrapError("Mkdir", name, err)
	}
	// Set the permissions after creation
	return wrapError("Mkdir.Chmod", name, fs.client.Chmod(name, perm))
}

// Remove removes a file or empty directory from the SFTP server.
func (fs *FileSystem) Remove(name string) error {
	return wrapError("Remove", name, fs.client.Remove(name))
}

// Rename renames a file on the SFTP server.
func (fs *FileSystem) Rename(oldpath, newpath string) error {
	err := fs.client.Rename(oldpath, newpath)
	if err != nil {
		return wrapErrorf("Rename(%s -> %s): %w", oldpath, newpath, err)
	}
	return nil
}

// Stat returns file info for a file on the SFTP server.
func (fs *FileSystem) Stat(name string) (os.FileInfo, error) {
	info, err := fs.client.Stat(name)
	if err != nil {
		return nil, wrapError("Stat", name, err)
	}
	return info, nil
}

// Chmod changes the mode of a file on the SFTP server.
func (fs *FileSystem) Chmod(name string, mode os.FileMode) error {
	return wrapError("Chmod", name, fs.client.Chmod(name, mode))
}

// Chtimes changes the access and modification times of a file on the SFTP server.
func (fs *FileSystem) Chtimes(name string, atime time.Time, mtime time.Time) error {
	return wrapError("Chtimes", name, fs.client.Chtimes(name, atime, mtime))
}

// Chown changes the owner and group of a file on the SFTP server.
func (fs *FileSystem) Chown(name string, uid, gid int) error {
	return wrapError("Chown", name, fs.client.Chown(name, uid, gid))
}

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

// MkdirAll creates a directory named path, along with any necessary parents.
func (fs *FileSystem) MkdirAll(path string, perm os.FileMode) error {
	// Walk up the path and create each directory
	return wrapError("MkdirAll", path, fs.client.MkdirAll(path))
}

// RemoveAll removes path and any children it contains.
func (fs *FileSystem) RemoveAll(path string) error {
	// First check if path exists and what it is
	info, err := fs.client.Stat(path)
	if err != nil {
		return wrapError("RemoveAll.Stat", path, err)
	}

	if !info.IsDir() {
		// If it's a file, just remove it
		return wrapError("RemoveAll", path, fs.client.Remove(path))
	}

	// For directories, we need to recursively remove contents
	return wrapError("RemoveAll", path, fs.removeAllDir(path))
}

// removeAllDir recursively removes a directory and all its contents
func (fs *FileSystem) removeAllDir(path string) error {
	// Read directory contents
	entries, err := fs.client.ReadDir(path)
	if err != nil {
		return err
	}

	// Remove each entry
	for _, entry := range entries {
		fullPath := path + "/" + entry.Name()
		if entry.IsDir() {
			// Recursively remove subdirectory
			if err := fs.removeAllDir(fullPath); err != nil {
				return err
			}
		} else {
			// Remove file
			if err := fs.client.Remove(fullPath); err != nil {
				return err
			}
		}
	}

	// Finally remove the directory itself
	return fs.client.Remove(path)
}

// Open opens the named file for reading.
func (fs *FileSystem) Open(name string) (absfs.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

// Create creates or truncates the named file.
func (fs *FileSystem) Create(name string) (absfs.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

// Truncate changes the size of the named file.
func (fs *FileSystem) Truncate(name string, size int64) error {
	return wrapError("Truncate", name, fs.client.Truncate(name, size))
}

// Separator returns the path separator for the filesystem.
func (fs *FileSystem) Separator() uint8 {
	return '/' // SFTP always uses forward slash
}

// ListSeparator returns the list separator for the filesystem.
func (fs *FileSystem) ListSeparator() uint8 {
	return ':' // Standard Unix path list separator
}

// Chdir changes the current working directory.
func (fs *FileSystem) Chdir(dir string) error {
	// Verify the directory exists
	info, err := fs.client.Stat(dir)
	if err != nil {
		return wrapError("Chdir.Stat", dir, err)
	}
	if !info.IsDir() {
		return wrapError("Chdir", dir, os.ErrInvalid)
	}
	fs.cwd = dir
	return nil
}

// Getwd returns the current working directory.
func (fs *FileSystem) Getwd() (string, error) {
	return fs.cwd, nil
}

// TempDir returns the temporary directory path.
func (fs *FileSystem) TempDir() string {
	return "/tmp" // Standard Unix temp directory
}

// Lstat returns file info without following symbolic links.
func (fs *FileSystem) Lstat(name string) (os.FileInfo, error) {
	info, err := fs.client.Lstat(name)
	if err != nil {
		return nil, wrapError("Lstat", name, err)
	}
	return info, nil
}

// Lchown changes the owner and group of a file without following symbolic links.
func (fs *FileSystem) Lchown(name string, uid, gid int) error {
	// SFTP doesn't have a direct Lchown operation, but we can use Chown
	// since it operates on the link itself when the target is a symlink
	return wrapError("Lchown", name, fs.client.Chown(name, uid, gid))
}

// Readlink returns the destination of a symbolic link.
func (fs *FileSystem) Readlink(name string) (string, error) {
	target, err := fs.client.ReadLink(name)
	if err != nil {
		return "", wrapError("Readlink", name, err)
	}
	return target, nil
}

// Symlink creates a symbolic link.
func (fs *FileSystem) Symlink(oldname, newname string) error {
	err := fs.client.Symlink(oldname, newname)
	if err != nil {
		return wrapErrorf("Symlink(%s -> %s): %w", oldname, newname, err)
	}
	return nil
}

// Compile-time verification that FileSystem implements absfs interfaces
var _ absfs.Filer = (*FileSystem)(nil)
var _ absfs.FileSystem = (*FileSystem)(nil)
var _ absfs.SymLinker = (*FileSystem)(nil)
var _ absfs.SymlinkFileSystem = (*FileSystem)(nil)
