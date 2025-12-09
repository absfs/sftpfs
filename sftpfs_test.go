package sftpfs

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/absfs/sftpfs/internal/mocks"
)

// mockSFTPClient is a test double for sftpClientInterface.
type mockSFTPClient struct {
	files       map[string]*mocks.MockSFTPFile
	dirs        map[string][]os.FileInfo
	fileInfos   map[string]os.FileInfo
	closeErr    error
	openFileErr error
	mkdirErr    error
	removeErr   error
	renameErr   error
	statErr     error
	chmodErr    error
	chtimesErr  error
	chownErr    error
	readDirErr  error
	closed      bool
}

func newMockSFTPClient() *mockSFTPClient {
	return &mockSFTPClient{
		files:     make(map[string]*mocks.MockSFTPFile),
		dirs:      make(map[string][]os.FileInfo),
		fileInfos: make(map[string]os.FileInfo),
	}
}

func (c *mockSFTPClient) Close() error {
	c.closed = true
	return c.closeErr
}

func (c *mockSFTPClient) OpenFile(path string, f int) (sftpFileInterface, error) {
	if c.openFileErr != nil {
		return nil, c.openFileErr
	}
	file, ok := c.files[path]
	if !ok {
		// Create new file for write operations
		if f&os.O_CREATE != 0 || f&os.O_WRONLY != 0 || f&os.O_RDWR != 0 {
			file = &mocks.MockSFTPFile{Data: []byte{}}
			c.files[path] = file
		} else {
			return nil, os.ErrNotExist
		}
	}
	// Reset position for new open
	file.Position = 0
	file.Closed = false
	return file, nil
}

func (c *mockSFTPClient) Mkdir(path string) error {
	if c.mkdirErr != nil {
		return c.mkdirErr
	}
	if _, exists := c.dirs[path]; exists {
		return os.ErrExist
	}
	c.dirs[path] = []os.FileInfo{}
	return nil
}

func (c *mockSFTPClient) Remove(path string) error {
	if c.removeErr != nil {
		return c.removeErr
	}
	if _, ok := c.files[path]; ok {
		delete(c.files, path)
		return nil
	}
	if _, ok := c.dirs[path]; ok {
		delete(c.dirs, path)
		return nil
	}
	return os.ErrNotExist
}

func (c *mockSFTPClient) Rename(oldpath, newpath string) error {
	if c.renameErr != nil {
		return c.renameErr
	}
	if file, ok := c.files[oldpath]; ok {
		c.files[newpath] = file
		delete(c.files, oldpath)
		return nil
	}
	return os.ErrNotExist
}

func (c *mockSFTPClient) Stat(path string) (os.FileInfo, error) {
	if c.statErr != nil {
		return nil, c.statErr
	}
	if info, ok := c.fileInfos[path]; ok {
		return info, nil
	}
	if file, ok := c.files[path]; ok {
		return &mocks.MockFileInfo{
			FileName: path,
			FileSize: int64(len(file.Data)),
			FileMode: 0644,
		}, nil
	}
	if _, ok := c.dirs[path]; ok {
		return &mocks.MockFileInfo{
			FileName:  path,
			FileIsDir: true,
			FileMode:  os.ModeDir | 0755,
		}, nil
	}
	return nil, os.ErrNotExist
}

func (c *mockSFTPClient) Chmod(path string, mode os.FileMode) error {
	if c.chmodErr != nil {
		return c.chmodErr
	}
	if _, ok := c.files[path]; !ok {
		if _, ok := c.dirs[path]; !ok {
			return os.ErrNotExist
		}
	}
	return nil
}

func (c *mockSFTPClient) Chtimes(path string, atime, mtime time.Time) error {
	if c.chtimesErr != nil {
		return c.chtimesErr
	}
	if _, ok := c.files[path]; !ok {
		if _, ok := c.dirs[path]; !ok {
			return os.ErrNotExist
		}
	}
	return nil
}

func (c *mockSFTPClient) Chown(path string, uid, gid int) error {
	if c.chownErr != nil {
		return c.chownErr
	}
	if _, ok := c.files[path]; !ok {
		if _, ok := c.dirs[path]; !ok {
			return os.ErrNotExist
		}
	}
	return nil
}

func (c *mockSFTPClient) ReadDir(path string) ([]os.FileInfo, error) {
	if c.readDirErr != nil {
		return nil, c.readDirErr
	}
	if entries, ok := c.dirs[path]; ok {
		return entries, nil
	}
	return nil, os.ErrNotExist
}

// Tests for Config struct
func TestConfig(t *testing.T) {
	config := &Config{
		Host:     "localhost:22",
		User:     "testuser",
		Password: "testpass",
		Timeout:  30 * time.Second,
	}

	if config.Host != "localhost:22" {
		t.Errorf("Host not set correctly")
	}
	if config.User != "testuser" {
		t.Errorf("User not set correctly")
	}
	if config.Password != "testpass" {
		t.Errorf("Password not set correctly")
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout not set correctly")
	}
}

func TestConfigWithKey(t *testing.T) {
	key := []byte("fake-private-key")
	config := &Config{
		Host: "localhost:22",
		User: "testuser",
		Key:  key,
	}

	if config.Host != "localhost:22" {
		t.Errorf("Host not set correctly")
	}
	if config.User != "testuser" {
		t.Errorf("User not set correctly")
	}
	if string(config.Key) != "fake-private-key" {
		t.Errorf("Key not set correctly")
	}
}

// Tests for New() function
func TestNewConnectionError(t *testing.T) {
	config := &Config{
		Host:     "localhost:22",
		User:     "testuser",
		Password: "testpass",
	}

	// Note: This will fail without an actual SFTP server
	_, err := New(config)
	if err == nil {
		t.Skip("Skipping test - SFTP server available unexpectedly")
	}
	// We expect an error since there's no server running
}

func TestNewDefaultTimeout(t *testing.T) {
	config := &Config{
		Host:     "localhost:22",
		User:     "testuser",
		Password: "testpass",
		Timeout:  0, // Should be set to default
	}

	// Try to create (will fail but should set default timeout)
	_, _ = New(config)

	if config.Timeout != 30*time.Second {
		t.Errorf("Default timeout not set, got %v", config.Timeout)
	}
}

func TestNewInvalidKey(t *testing.T) {
	config := &Config{
		Host: "localhost:22",
		User: "testuser",
		Key:  []byte("not-a-valid-key"),
	}

	_, err := New(config)
	if err == nil {
		t.Error("Expected error for invalid key")
	}
}

// Tests for Dial() function signature
func TestDialSignature(t *testing.T) {
	var _ func(string, string, string) (*FileSystem, error) = Dial
}

// Tests for DialWithKey() function signature
func TestDialWithKeySignature(t *testing.T) {
	var _ func(string, string, []byte) (*FileSystem, error) = DialWithKey
}

// Tests using mock clients
func TestNewWithClients(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockSSH := &mocks.MockSSHClient{}

	fs := newWithClients(mockClient, mockSSH)

	if fs.client == nil {
		t.Error("Expected client to be set")
	}
	if fs.sshClient == nil {
		t.Error("Expected sshClient to be set")
	}
}

func TestClose(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockSSH := &mocks.MockSSHClient{}

	fs := newWithClients(mockClient, mockSSH)
	err := fs.Close()

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !mockClient.closed {
		t.Error("Expected SFTP client to be closed")
	}
	if !mockSSH.Closed {
		t.Error("Expected SSH client to be closed")
	}
}

func TestCloseWithError(t *testing.T) {
	expectedErr := errors.New("close error")
	mockClient := newMockSFTPClient()
	mockSSH := &mocks.MockSSHClient{CloseErr: expectedErr}

	fs := newWithClients(mockClient, mockSSH)
	err := fs.Close()

	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestCloseNilClients(t *testing.T) {
	fs := &FileSystem{
		client:    nil,
		sshClient: nil,
	}

	// Should not panic
	err := fs.Close()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}

// Tests for FileSystem file operations
func TestOpenFile(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{Data: []byte("hello")}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	file, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer file.Close()

	if file.Name() != "/test.txt" {
		t.Errorf("Expected name /test.txt, got %s", file.Name())
	}
}

func TestOpenFileCreate(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	file, err := fs.OpenFile("/new.txt", os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	defer file.Close()

	if file.Name() != "/new.txt" {
		t.Errorf("Expected name /new.txt, got %s", file.Name())
	}
}

func TestOpenFileError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.openFileErr = errors.New("open error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	_, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestOpenFileNotExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	_, err := fs.OpenFile("/nonexistent.txt", os.O_RDONLY, 0644)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}
}

func TestStat(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{Data: []byte("hello")}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	info, err := fs.Stat("/test.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if info.Name() != "/test.txt" {
		t.Errorf("Expected name /test.txt, got %s", info.Name())
	}
	if info.Size() != 5 {
		t.Errorf("Expected size 5, got %d", info.Size())
	}
}

func TestStatDirectory(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	info, err := fs.Stat("/testdir")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !info.IsDir() {
		t.Error("Expected directory")
	}
}

func TestStatError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.statErr = errors.New("stat error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	_, err := fs.Stat("/test.txt")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestStatNotExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	_, err := fs.Stat("/nonexistent.txt")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}
}

func TestRemove(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Remove("/test.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, exists := mockClient.files["/test.txt"]; exists {
		t.Error("Expected file to be removed")
	}
}

func TestRemoveDirectory(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Remove("/testdir")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, exists := mockClient.dirs["/testdir"]; exists {
		t.Error("Expected directory to be removed")
	}
}

func TestRemoveError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.removeErr = errors.New("remove error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Remove("/test.txt")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestRemoveNotExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Remove("/nonexistent.txt")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}
}

func TestRename(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/old.txt"] = &mocks.MockSFTPFile{Data: []byte("content")}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Rename("/old.txt", "/new.txt")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, exists := mockClient.files["/old.txt"]; exists {
		t.Error("Expected old file to be removed")
	}
	if _, exists := mockClient.files["/new.txt"]; !exists {
		t.Error("Expected new file to exist")
	}
}

func TestRenameError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.renameErr = errors.New("rename error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Rename("/old.txt", "/new.txt")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestRenameNotExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Rename("/nonexistent.txt", "/new.txt")
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}
}

func TestChmod(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chmod("/test.txt", 0755)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestChmodDirectory(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chmod("/testdir", 0755)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestChmodError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.chmodErr = errors.New("chmod error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chmod("/test.txt", 0755)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestChmodNotExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chmod("/nonexistent.txt", 0755)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}
}

func TestChtimes(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	now := time.Now()
	err := fs.Chtimes("/test.txt", now, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestChtimesDirectory(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	now := time.Now()
	err := fs.Chtimes("/testdir", now, now)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestChtimesError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.chtimesErr = errors.New("chtimes error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	now := time.Now()
	err := fs.Chtimes("/test.txt", now, now)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestChtimesNotExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	now := time.Now()
	err := fs.Chtimes("/nonexistent.txt", now, now)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}
}

func TestChown(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chown("/test.txt", 1000, 1000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestChownDirectory(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chown("/testdir", 1000, 1000)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestChownError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.chownErr = errors.New("chown error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chown("/test.txt", 1000, 1000)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestChownNotExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Chown("/nonexistent.txt", 1000, 1000)
	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("Expected ErrNotExist, got %v", err)
	}
}

// Tests for directory operations
func TestMkdir(t *testing.T) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Mkdir("/newdir", 0755)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if _, exists := mockClient.dirs["/newdir"]; !exists {
		t.Error("Expected directory to be created")
	}
}

func TestMkdirAlreadyExists(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/existingdir"] = []os.FileInfo{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Mkdir("/existingdir", 0755)
	if !errors.Is(err, os.ErrExist) {
		t.Errorf("Expected ErrExist, got %v", err)
	}
}

func TestMkdirError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.mkdirErr = errors.New("mkdir error")

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	err := fs.Mkdir("/newdir", 0755)
	if err == nil {
		t.Error("Expected error")
	}
}

// Tests for File wrapper methods
func TestFileName(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{}
	file := &File{
		file: mockFile,
		name: "/test.txt",
	}

	if file.Name() != "/test.txt" {
		t.Errorf("Expected /test.txt, got %s", file.Name())
	}
}

func TestFileRead(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello world")}
	file := &File{file: mockFile, name: "/test.txt"}

	buf := make([]byte, 5)
	n, err := file.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(buf))
	}
}

func TestFileReadAll(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello")}
	file := &File{file: mockFile, name: "/test.txt"}

	buf := make([]byte, 10)
	n, err := file.Read(buf)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}

	// Read again should return EOF
	n, err = file.Read(buf)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 bytes, got %d", n)
	}
}

func TestFileReadError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{ReadErr: errors.New("read error")}
	file := &File{file: mockFile, name: "/test.txt"}

	buf := make([]byte, 5)
	_, err := file.Read(buf)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileReadAt(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello world")}
	file := &File{file: mockFile, name: "/test.txt"}

	buf := make([]byte, 5)
	n, err := file.ReadAt(buf, 6)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != "world" {
		t.Errorf("Expected 'world', got '%s'", string(buf))
	}
}

func TestFileReadAtError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{ReadErr: errors.New("read error")}
	file := &File{file: mockFile, name: "/test.txt"}

	buf := make([]byte, 5)
	_, err := file.ReadAt(buf, 0)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileReadAtEOF(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello")}
	file := &File{file: mockFile, name: "/test.txt"}

	buf := make([]byte, 5)
	_, err := file.ReadAt(buf, 100)
	if err != io.EOF {
		t.Errorf("Expected EOF, got %v", err)
	}
}

func TestFileWrite(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte{}}
	file := &File{file: mockFile, name: "/test.txt"}

	n, err := file.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(mockFile.Data) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(mockFile.Data))
	}
}

func TestFileWriteError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{WriteErr: errors.New("write error")}
	file := &File{file: mockFile, name: "/test.txt"}

	_, err := file.Write([]byte("hello"))
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileWriteAt(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: make([]byte, 10)}
	file := &File{file: mockFile, name: "/test.txt"}

	n, err := file.WriteAt([]byte("hello"), 5)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
}

func TestFileWriteAtError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{WriteErr: errors.New("write error")}
	file := &File{file: mockFile, name: "/test.txt"}

	_, err := file.WriteAt([]byte("hello"), 0)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileWriteString(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte{}}
	file := &File{file: mockFile, name: "/test.txt"}

	n, err := file.WriteString("hello")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(mockFile.Data) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(mockFile.Data))
	}
}

func TestFileSeekStart(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello world")}
	file := &File{file: mockFile, name: "/test.txt"}

	pos, err := file.Seek(5, io.SeekStart)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pos != 5 {
		t.Errorf("Expected position 5, got %d", pos)
	}
}

func TestFileSeekCurrent(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello world"), Position: 5}
	file := &File{file: mockFile, name: "/test.txt"}

	pos, err := file.Seek(3, io.SeekCurrent)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pos != 8 {
		t.Errorf("Expected position 8, got %d", pos)
	}
}

func TestFileSeekEnd(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello world")}
	file := &File{file: mockFile, name: "/test.txt"}

	pos, err := file.Seek(-5, io.SeekEnd)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if pos != 6 {
		t.Errorf("Expected position 6, got %d", pos)
	}
}

func TestFileSeekError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{SeekErr: errors.New("seek error")}
	file := &File{file: mockFile, name: "/test.txt"}

	_, err := file.Seek(0, io.SeekStart)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileSeekNegative(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello")}
	file := &File{file: mockFile, name: "/test.txt"}

	_, err := file.Seek(-10, io.SeekStart)
	if !errors.Is(err, fs.ErrInvalid) {
		t.Errorf("Expected ErrInvalid, got %v", err)
	}
}

func TestFileClose(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{}
	file := &File{file: mockFile, name: "/test.txt"}

	err := file.Close()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if !mockFile.Closed {
		t.Error("Expected file to be closed")
	}
}

func TestFileCloseError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{CloseErr: errors.New("close error")}
	file := &File{file: mockFile, name: "/test.txt"}

	err := file.Close()
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileStat(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{
		Data: []byte("hello"),
		StatInfo: &mocks.MockFileInfo{
			FileName: "test.txt",
			FileSize: 5,
			FileMode: 0644,
		},
	}
	file := &File{file: mockFile, name: "/test.txt"}

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if info.Name() != "test.txt" {
		t.Errorf("Expected name test.txt, got %s", info.Name())
	}
	if info.Size() != 5 {
		t.Errorf("Expected size 5, got %d", info.Size())
	}
}

func TestFileStatDefault(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello")}
	file := &File{file: mockFile, name: "/test.txt"}

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if info.Size() != 5 {
		t.Errorf("Expected size 5, got %d", info.Size())
	}
}

func TestFileStatError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{StatErr: errors.New("stat error")}
	file := &File{file: mockFile, name: "/test.txt"}

	_, err := file.Stat()
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileSync(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{}
	file := &File{file: mockFile, name: "/test.txt"}

	// Sync should be a no-op
	err := file.Sync()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func TestFileTruncate(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello world")}
	file := &File{file: mockFile, name: "/test.txt"}

	err := file.Truncate(5)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if string(mockFile.Data) != "hello" {
		t.Errorf("Expected 'hello', got '%s'", string(mockFile.Data))
	}
}

func TestFileTruncateExpand(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{Data: []byte("hello")}
	file := &File{file: mockFile, name: "/test.txt"}

	err := file.Truncate(10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(mockFile.Data) != 10 {
		t.Errorf("Expected length 10, got %d", len(mockFile.Data))
	}
}

func TestFileTruncateError(t *testing.T) {
	mockFile := &mocks.MockSFTPFile{TruncateErr: errors.New("truncate error")}
	file := &File{file: mockFile, name: "/test.txt"}

	err := file.Truncate(5)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileReaddir(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{
		&mocks.MockFileInfo{FileName: "file1.txt"},
		&mocks.MockFileInfo{FileName: "file2.txt"},
		&mocks.MockFileInfo{FileName: "file3.txt"},
	}

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	entries, err := file.Readdir(-1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestFileReaddirLimited(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{
		&mocks.MockFileInfo{FileName: "file1.txt"},
		&mocks.MockFileInfo{FileName: "file2.txt"},
		&mocks.MockFileInfo{FileName: "file3.txt"},
	}

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	entries, err := file.Readdir(2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestFileReaddirMoreThanAvailable(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{
		&mocks.MockFileInfo{FileName: "file1.txt"},
	}

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	entries, err := file.Readdir(10)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
}

func TestFileReaddirZero(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{
		&mocks.MockFileInfo{FileName: "file1.txt"},
		&mocks.MockFileInfo{FileName: "file2.txt"},
	}

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	// n=0 should return all entries
	entries, err := file.Readdir(0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}
}

func TestFileReaddirError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.readDirErr = errors.New("readdir error")

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	_, err := file.Readdir(-1)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestFileReaddirnames(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{
		&mocks.MockFileInfo{FileName: "file1.txt"},
		&mocks.MockFileInfo{FileName: "file2.txt"},
		&mocks.MockFileInfo{FileName: "file3.txt"},
	}

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	names, err := file.Readdirnames(-1)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}
	if names[0] != "file1.txt" {
		t.Errorf("Expected file1.txt, got %s", names[0])
	}
}

func TestFileReaddirnamesLimited(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.dirs["/testdir"] = []os.FileInfo{
		&mocks.MockFileInfo{FileName: "file1.txt"},
		&mocks.MockFileInfo{FileName: "file2.txt"},
		&mocks.MockFileInfo{FileName: "file3.txt"},
	}

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	names, err := file.Readdirnames(2)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
}

func TestFileReaddirnamesError(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.readDirErr = errors.New("readdir error")

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	_, err := file.Readdirnames(-1)
	if err == nil {
		t.Error("Expected error")
	}
}

// Additional coverage tests for Dial and DialWithKey convenience functions
func TestDialIntegration(t *testing.T) {
	// Test Dial function - will fail without server, which is expected
	_, err := Dial("nonexistent.invalid:22", "user", "pass")
	if err == nil {
		t.Skip("Unexpected connection - SFTP server available")
	}
	// We expect an error since there's no server
}

func TestDialWithKeyIntegration(t *testing.T) {
	// Test DialWithKey function - will fail without server, which is expected
	fakeKey := []byte("not-a-real-key")
	_, err := DialWithKey("nonexistent.invalid:22", "user", fakeKey)
	if err == nil {
		t.Skip("Unexpected connection - SFTP server available")
	}
	// We expect an error since the key is invalid or there's no server
}

// Test for New() function with password to ensure that path is covered
func TestNewWithPassword(t *testing.T) {
	config := &Config{
		Host:     "nonexistent.invalid:22",
		User:     "testuser",
		Password: "testpass",
		Timeout:  1 * time.Second,
	}

	// This will fail to connect, but ensures the password auth path is tested
	_, err := New(config)
	if err == nil {
		t.Skip("Unexpected connection - SFTP server available")
	}
	// We expect a connection error
}

// Test for New() function with explicit timeout to cover that branch
func TestNewWithExplicitTimeout(t *testing.T) {
	config := &Config{
		Host:     "nonexistent.invalid:22",
		User:     "testuser",
		Password: "testpass",
		Timeout:  5 * time.Second, // Explicit non-zero timeout
	}

	_, err := New(config)
	if err == nil {
		t.Skip("Unexpected connection - SFTP server available")
	}
	// Verify the timeout was not changed
	if config.Timeout != 5*time.Second {
		t.Errorf("Expected timeout to remain 5s, got %v", config.Timeout)
	}
}
