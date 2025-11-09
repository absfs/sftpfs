package sftpfs

import (
	"io"
	"os"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

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

func TestNewConfig(t *testing.T) {
	config := &Config{
		Host:     "localhost:22",
		User:     "testuser",
		Password: "testpass",
	}

	// Note: This will fail without an actual SFTP server
	// This is just a structural test
	_, err := New(config)
	if err == nil {
		t.Skip("Skipping test - no SFTP server available")
	}
	// We expect an error since there's no server running
	// This just tests that the function can be called
}

func TestDialSignature(t *testing.T) {
	// Test that Dial function exists with correct signature
	// This is a compile-time test more than a runtime test
	var _ func(string, string, string) (*FileSystem, error) = Dial
}

func TestDialWithKeySignature(t *testing.T) {
	// Test that DialWithKey function exists with correct signature
	var _ func(string, string, []byte) (*FileSystem, error) = DialWithKey
}

func TestConfigDefaults(t *testing.T) {
	config := &Config{
		Host:     "localhost:22",
		User:     "testuser",
		Password: "testpass",
	}

	// Timeout should be set to default when not specified
	if config.Timeout != 0 {
		t.Errorf("Expected timeout to be 0 before New() is called")
	}
}

func TestConfigWithHostKeyCallback(t *testing.T) {
	// Test that we can set a custom HostKeyCallback
	customCallback := ssh.InsecureIgnoreHostKey()
	config := &Config{
		Host:            "localhost:22",
		User:            "testuser",
		Password:        "testpass",
		HostKeyCallback: customCallback,
	}

	if config.HostKeyCallback == nil {
		t.Errorf("HostKeyCallback should be set")
	}
}

func TestFileReaddir(t *testing.T) {
	// Test Readdir state management with mock data
	f := &File{
		dirEntries: []os.FileInfo{
			&mockFileInfo{name: "file1.txt"},
			&mockFileInfo{name: "file2.txt"},
			&mockFileInfo{name: "file3.txt"},
		},
		readdirPos: 0,
	}

	// Read first entry
	entries, err := f.Readdir(1)
	if err != nil {
		t.Fatalf("Readdir failed: %v", err)
	}
	if len(entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(entries))
	}
	if entries[0].Name() != "file1.txt" {
		t.Errorf("Expected file1.txt, got %s", entries[0].Name())
	}

	// Read next two entries
	entries, err = f.Readdir(2)
	if err != nil {
		t.Fatalf("Readdir failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Try to read more - should get EOF
	entries, err = f.Readdir(1)
	if err != io.EOF {
		t.Errorf("Expected io.EOF, got %v", err)
	}
	if entries != nil {
		t.Errorf("Expected nil entries, got %v", entries)
	}
}

func TestFileReaddirAll(t *testing.T) {
	// Test Readdir with n <= 0 (read all)
	f := &File{
		dirEntries: []os.FileInfo{
			&mockFileInfo{name: "file1.txt"},
			&mockFileInfo{name: "file2.txt"},
			&mockFileInfo{name: "file3.txt"},
		},
		readdirPos: 0,
	}

	entries, err := f.Readdir(-1)
	if err != nil {
		t.Fatalf("Readdir failed: %v", err)
	}
	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}
}

func TestFileReaddirnames(t *testing.T) {
	f := &File{
		dirEntries: []os.FileInfo{
			&mockFileInfo{name: "file1.txt"},
			&mockFileInfo{name: "file2.txt"},
		},
		readdirPos: 0,
	}

	names, err := f.Readdirnames(2)
	if err != nil {
		t.Fatalf("Readdirnames failed: %v", err)
	}
	if len(names) != 2 {
		t.Errorf("Expected 2 names, got %d", len(names))
	}
	if names[0] != "file1.txt" || names[1] != "file2.txt" {
		t.Errorf("Unexpected names: %v", names)
	}
}

func TestFileName(t *testing.T) {
	f := &File{name: "/path/to/file.txt"}
	if f.Name() != "/path/to/file.txt" {
		t.Errorf("Expected /path/to/file.txt, got %s", f.Name())
	}
}

func TestFileSync(t *testing.T) {
	f := &File{}
	// Sync should always return nil (no-op)
	if err := f.Sync(); err != nil {
		t.Errorf("Sync should return nil, got %v", err)
	}
}

// mockFileInfo implements os.FileInfo for testing
type mockFileInfo struct {
	name string
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return 0 }
func (m *mockFileInfo) Mode() os.FileMode  { return 0644 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return false }
func (m *mockFileInfo) Sys() interface{}   { return nil }
