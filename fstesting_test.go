package sftpfs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/absfs/absfs"
	"github.com/absfs/sftpfs/internal/mocks"
)

// enhancedMockSFTPClient extends mockSFTPClient with additional functionality
// needed for fstesting Suite tests.
type enhancedMockSFTPClient struct {
	*mockSFTPClient
	permissions map[string]os.FileMode
	times       map[string]time.Time
}

func newEnhancedMockSFTPClient() *enhancedMockSFTPClient {
	return &enhancedMockSFTPClient{
		mockSFTPClient: newMockSFTPClient(),
		permissions:    make(map[string]os.FileMode),
		times:          make(map[string]time.Time),
	}
}

func (c *enhancedMockSFTPClient) OpenFile(path string, f int) (sftpFileInterface, error) {
	// Normalize path (remove trailing slash)
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	// Check if path is a directory
	if _, isDir := c.dirs[path]; isDir {
		// Allow opening directories for reading only
		if f&os.O_WRONLY != 0 || f&os.O_RDWR != 0 {
			return nil, os.ErrInvalid
		}
		// Return a special file handle for directories
		return &mocks.MockSFTPFile{Data: []byte{}}, nil
	}

	// Handle O_CREATE and O_EXCL flags
	if f&os.O_EXCL != 0 {
		if _, ok := c.files[path]; ok {
			return nil, os.ErrExist
		}
	}

	// Handle O_TRUNC flag
	if f&os.O_TRUNC != 0 {
		if file, ok := c.files[path]; ok {
			file.Data = []byte{}
			file.Position = 0
		}
	}

	// Call parent implementation
	file, err := c.mockSFTPClient.OpenFile(path, f)
	if err != nil {
		return nil, err
	}

	// Handle O_APPEND flag
	if f&os.O_APPEND != 0 {
		if mockFile, ok := file.(*mocks.MockSFTPFile); ok {
			mockFile.Position = int64(len(mockFile.Data))
		}
	}

	// Create directory entry if creating a new file
	if f&os.O_CREATE != 0 {
		dir := filepath.Dir(path)
		if dir != "." && dir != "/" {
			c.ensureDirExists(dir)
		}
	}

	return file, nil
}

func (c *enhancedMockSFTPClient) Mkdir(path string) error {
	err := c.mockSFTPClient.Mkdir(path)
	if err != nil {
		return err
	}

	// Ensure parent directory exists
	parent := filepath.Dir(path)
	if parent != "." && parent != "/" {
		c.ensureDirExists(parent)
	}

	// Set default permissions
	c.permissions[path] = 0755
	c.times[path] = time.Now()

	return nil
}

func (c *enhancedMockSFTPClient) Remove(path string) error {
	err := c.mockSFTPClient.Remove(path)
	if err != nil {
		return err
	}

	delete(c.permissions, path)
	delete(c.times, path)
	return nil
}

func (c *enhancedMockSFTPClient) Chmod(path string, mode os.FileMode) error {
	err := c.mockSFTPClient.Chmod(path, mode)
	if err != nil {
		return err
	}
	c.permissions[path] = mode
	return nil
}

func (c *enhancedMockSFTPClient) Chtimes(path string, atime, mtime time.Time) error {
	err := c.mockSFTPClient.Chtimes(path, atime, mtime)
	if err != nil {
		return err
	}
	c.times[path] = mtime
	return nil
}

func (c *enhancedMockSFTPClient) ReadDir(path string) ([]os.FileInfo, error) {
	if c.readDirErr != nil {
		return nil, c.readDirErr
	}

	// Normalize path
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	// Check if directory exists
	if _, ok := c.dirs[path]; !ok {
		return nil, os.ErrNotExist
	}

	// Build list of entries in this directory
	var entries []os.FileInfo

	// Find all files in this directory
	for filePath, file := range c.files {
		dir := filepath.Dir(filePath)
		if dir == path {
			mode := c.permissions[filePath]
			if mode == 0 {
				mode = 0644
			}
			modTime := c.times[filePath]
			if modTime.IsZero() {
				modTime = time.Now()
			}
			entries = append(entries, &mocks.MockFileInfo{
				FileName:    filepath.Base(filePath),
				FileSize:    int64(len(file.Data)),
				FileMode:    mode,
				FileModTime: modTime,
				FileIsDir:   false,
			})
		}
	}

	// Find all subdirectories in this directory
	for dirPath := range c.dirs {
		if dirPath == path || dirPath == "/" {
			continue
		}
		parent := filepath.Dir(dirPath)
		if parent == path {
			mode := c.permissions[dirPath]
			if mode == 0 {
				mode = os.ModeDir | 0755
			} else if mode&os.ModeDir == 0 {
				mode = os.ModeDir | mode
			}
			modTime := c.times[dirPath]
			if modTime.IsZero() {
				modTime = time.Now()
			}
			entries = append(entries, &mocks.MockFileInfo{
				FileName:    filepath.Base(dirPath),
				FileIsDir:   true,
				FileMode:    mode,
				FileModTime: modTime,
			})
		}
	}

	return entries, nil
}

func (c *enhancedMockSFTPClient) Stat(path string) (os.FileInfo, error) {
	if c.statErr != nil {
		return nil, c.statErr
	}

	// Normalize path (remove trailing slash except for root)
	originalPath := path
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	// Check for custom file info first
	if info, ok := c.fileInfos[path]; ok {
		return info, nil
	}

	// Check for file
	if file, ok := c.files[path]; ok {
		mode := c.permissions[path]
		if mode == 0 {
			mode = 0644
		}
		modTime := c.times[path]
		if modTime.IsZero() {
			modTime = time.Now()
		}

		baseName := filepath.Base(path)
		if originalPath != path && strings.HasSuffix(originalPath, "/") {
			baseName = filepath.Base(originalPath)
		}

		return &mocks.MockFileInfo{
			FileName:    baseName,
			FileSize:    int64(len(file.Data)),
			FileMode:    mode,
			FileModTime: modTime,
			FileIsDir:   false,
		}, nil
	}

	// Check for directory
	if _, ok := c.dirs[path]; ok {
		mode := c.permissions[path]
		if mode == 0 {
			mode = os.ModeDir | 0755
		} else if mode&os.ModeDir == 0 {
			mode = os.ModeDir | mode
		}
		modTime := c.times[path]
		if modTime.IsZero() {
			modTime = time.Now()
		}

		baseName := filepath.Base(path)
		if originalPath != path && strings.HasSuffix(originalPath, "/") {
			baseName = filepath.Base(strings.TrimSuffix(originalPath, "/"))
		}

		return &mocks.MockFileInfo{
			FileName:    baseName,
			FileIsDir:   true,
			FileMode:    mode,
			FileModTime: modTime,
		}, nil
	}

	return nil, os.ErrNotExist
}

func (c *enhancedMockSFTPClient) ensureDirExists(path string) {
	if path == "" || path == "." || path == "/" {
		return
	}

	// Ensure all parent directories exist
	parent := filepath.Dir(path)
	if parent != "." && parent != "/" {
		c.ensureDirExists(parent)
	}

	// Create this directory if it doesn't exist
	if _, exists := c.dirs[path]; !exists {
		c.dirs[path] = []os.FileInfo{}
		c.permissions[path] = os.ModeDir | 0755
		c.times[path] = time.Now()
	}
}

// mockFileSystemWrapper wraps the sftpfs with additional methods needed for fstesting.
type mockFileSystemWrapper struct {
	*FileSystem
	client *enhancedMockSFTPClient
}

func (fs *mockFileSystemWrapper) Separator() uint8 {
	return '/'
}

func (fs *mockFileSystemWrapper) ListSeparator() uint8 {
	return ':'
}

func (fs *mockFileSystemWrapper) Chdir(dir string) error {
	// SFTP doesn't support changing directories
	return nil
}

func (fs *mockFileSystemWrapper) Getwd() (string, error) {
	return "/", nil
}

func (fs *mockFileSystemWrapper) TempDir() string {
	return "/tmp"
}

func (fs *mockFileSystemWrapper) Open(name string) (absfs.File, error) {
	return fs.OpenFile(name, os.O_RDONLY, 0)
}

func (fs *mockFileSystemWrapper) Create(name string) (absfs.File, error) {
	return fs.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func (fs *mockFileSystemWrapper) MkdirAll(path string, perm os.FileMode) error {
	if path == "" || path == "." || path == "/" {
		return nil
	}

	// Check if already exists
	info, err := fs.Stat(path)
	if err == nil {
		if !info.IsDir() {
			return os.ErrExist
		}
		return nil
	}

	// Create parent first
	parent := filepath.Dir(path)
	if parent != "." && parent != "/" {
		if err := fs.MkdirAll(parent, perm); err != nil {
			return err
		}
	}

	// Create this directory
	return fs.Mkdir(path, perm)
}

func (fs *mockFileSystemWrapper) RemoveAll(path string) error {
	// Check if path exists
	info, err := fs.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	// If it's a file, just remove it
	if !info.IsDir() {
		return fs.Remove(path)
	}

	// For directories, we need to remove all contents first
	// In a real implementation, we'd recursively read and delete
	// For our mock, we'll just remove the directory
	return fs.Remove(path)
}

func (fs *mockFileSystemWrapper) Truncate(name string, size int64) error {
	file, err := fs.OpenFile(name, os.O_WRONLY, 0)
	if err != nil {
		return err
	}
	defer file.Close()

	return file.Truncate(size)
}

// createMockSFTPFS creates a mock SFTP filesystem for testing.
func createMockSFTPFS() absfs.FileSystem {
	client := newEnhancedMockSFTPClient()
	sshClient := &mocks.MockSSHClient{}

	// Create root and tmp directories
	client.dirs["/"] = []os.FileInfo{}
	client.dirs["/tmp"] = []os.FileInfo{}
	client.permissions["/"] = os.ModeDir | 0755
	client.permissions["/tmp"] = os.ModeDir | 0755

	fs := newWithClients(client, sshClient)

	return &mockFileSystemWrapper{
		FileSystem: fs,
		client:     client,
	}
}

// Note: The fstesting package API has been updated. The old Suite/Features API
// has been removed. Tests using the new API should be added here.

// TestCapabilities documents the capabilities of sftpfs.
func TestCapabilities(t *testing.T) {
	t.Log("SFTP FileSystem Capabilities:")
	t.Log("  - Symlinks: NO (requires extended SFTP protocol)")
	t.Log("  - Hard Links: NO (not supported by SFTP)")
	t.Log("  - Permissions: YES (Unix-style permissions)")
	t.Log("  - Timestamps: YES (atime/mtime)")
	t.Log("  - Case Sensitive: YES (Unix-based servers)")
	t.Log("  - Atomic Rename: YES")
	t.Log("  - Sparse Files: DEPENDS (on remote filesystem)")
	t.Log("  - Large Files: YES")
	t.Log("  - Node Type: Adapter (adapts SFTP to absfs)")
}

// TestWithPathVariations tests various path formats.
func TestWithPathVariations(t *testing.T) {
	fs := createMockSFTPFS()

	testCases := []struct {
		name string
		path string
		want bool
	}{
		{"absolute path", "/tmp/test.txt", true},
		{"path with spaces", "/tmp/file with spaces.txt", true},
		{"path with dots", "/tmp/file.multiple.dots.txt", true},
		{"nested path", "/tmp/nested/deep/path/file.txt", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Ensure parent directory exists
			parent := filepath.Dir(tc.path)
			if !strings.Contains(parent, "nested") {
				fs.MkdirAll(parent, 0755)
			} else {
				// For nested paths, create each level
				parts := strings.Split(strings.Trim(parent, "/"), "/")
				current := "/"
				for _, part := range parts {
					current = filepath.Join(current, part)
					fs.MkdirAll(current, 0755)
				}
			}

			// Create file
			f, err := fs.Create(tc.path)
			if err != nil {
				t.Fatalf("Create failed: %v", err)
			}
			f.Write([]byte("test"))
			f.Close()

			// Verify
			info, err := fs.Stat(tc.path)
			if err != nil {
				t.Fatalf("Stat failed: %v", err)
			}
			if info.IsDir() {
				t.Error("Expected file, got directory")
			}

			// Cleanup
			fs.Remove(tc.path)
		})
	}
}
