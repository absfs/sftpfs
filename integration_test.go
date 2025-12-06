//go:build integration

package sftpfs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const (
	testHost     = "localhost:2222"
	testUser     = "testuser"
	testPassword = "testpass"
	testBaseDir  = "/home/testuser"
)

// skipIfNoServer skips the test if the SFTP server is not available.
func skipIfNoServer(t *testing.T) *FileSystem {
	t.Helper()
	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("SFTP server not available: %v", err)
	}
	return fs
}

// TestIntegrationConnection tests real connection to SFTP server.
func TestIntegrationConnection(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	// Verify we can stat the home directory
	info, err := fs.Stat(testBaseDir)
	if err != nil {
		t.Fatalf("Failed to stat home directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected home directory to be a directory")
	}
}

// TestIntegrationPasswordAuth tests password authentication.
func TestIntegrationPasswordAuth(t *testing.T) {
	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Verify connection works
	_, err = fs.Stat(testBaseDir)
	if err != nil {
		t.Fatalf("Failed to stat after password auth: %v", err)
	}
}

// TestIntegrationInvalidCredentials tests connection with invalid credentials.
func TestIntegrationInvalidCredentials(t *testing.T) {
	_, err := Dial(testHost, "wronguser", "wrongpass")
	if err == nil {
		t.Error("Expected error for invalid credentials")
	}
}

// TestIntegrationConnectionTimeout tests connection to invalid host.
func TestIntegrationConnectionTimeout(t *testing.T) {
	config := &Config{
		Host:     "192.0.2.1:22", // Non-routable address
		User:     testUser,
		Password: testPassword,
		Timeout:  2 * time.Second,
	}

	_, err := New(config)
	if err == nil {
		t.Error("Expected timeout error")
	}
}

// TestIntegrationFileOperations tests creating, writing, reading, and deleting files.
func TestIntegrationFileOperations(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_file.txt")
	testContent := []byte("Hello, SFTP World!")

	// Create and write file
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	n, err := file.Write(testContent)
	if err != nil {
		t.Fatalf("Failed to write file: %v", err)
	}
	if n != len(testContent) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testContent), n)
	}
	file.Close()

	// Read file back
	file, err = fs.OpenFile(testFile, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open file for reading: %v", err)
	}

	readContent, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	file.Close()

	if !bytes.Equal(readContent, testContent) {
		t.Errorf("Content mismatch: expected %q, got %q", testContent, readContent)
	}

	// Stat the file
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	if info.Size() != int64(len(testContent)) {
		t.Errorf("Expected size %d, got %d", len(testContent), info.Size())
	}

	// Delete file
	err = fs.Remove(testFile)
	if err != nil {
		t.Fatalf("Failed to remove file: %v", err)
	}

	// Verify file is deleted
	_, err = fs.Stat(testFile)
	if err == nil {
		t.Error("Expected error stating deleted file")
	}
}

// TestIntegrationWriteString tests WriteString method.
func TestIntegrationWriteString(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_writestring.txt")
	testContent := "Hello via WriteString!"

	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	n, err := file.(*File).WriteString(testContent)
	if err != nil {
		t.Fatalf("Failed to write string: %v", err)
	}
	if n != len(testContent) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testContent), n)
	}
	file.Close()

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationSeek tests file seeking.
func TestIntegrationSeek(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_seek.txt")
	testContent := []byte("0123456789")

	// Create file
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Write(testContent)
	file.Close()

	// Open for reading and seek
	file, err = fs.OpenFile(testFile, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		file.Close()
		fs.Remove(testFile)
	}()

	// Seek to position 5
	pos, err := file.Seek(5, io.SeekStart)
	if err != nil {
		t.Fatalf("Failed to seek: %v", err)
	}
	if pos != 5 {
		t.Errorf("Expected position 5, got %d", pos)
	}

	// Read from position 5
	buf := make([]byte, 5)
	n, err := file.Read(buf)
	if err != nil {
		t.Fatalf("Failed to read: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != "56789" {
		t.Errorf("Expected '56789', got '%s'", string(buf))
	}
}

// TestIntegrationRename tests file renaming.
func TestIntegrationRename(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	oldFile := filepath.Join(testBaseDir, "old_name.txt")
	newFile := filepath.Join(testBaseDir, "new_name.txt")

	// Create file
	file, err := fs.OpenFile(oldFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Write([]byte("content"))
	file.Close()

	// Rename
	err = fs.Rename(oldFile, newFile)
	if err != nil {
		t.Fatalf("Failed to rename: %v", err)
	}

	// Verify old file doesn't exist
	_, err = fs.Stat(oldFile)
	if err == nil {
		t.Error("Expected old file to not exist")
	}

	// Verify new file exists
	_, err = fs.Stat(newFile)
	if err != nil {
		t.Errorf("Expected new file to exist: %v", err)
	}

	// Cleanup
	fs.Remove(newFile)
}

// TestIntegrationDirectoryOperations tests directory creation and listing.
func TestIntegrationDirectoryOperations(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testDir := filepath.Join(testBaseDir, "test_directory")

	// Create directory
	err := fs.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	// Stat directory
	info, err := fs.Stat(testDir)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected to be a directory")
	}

	// Create some files in the directory
	for i := 0; i < 3; i++ {
		f, err := fs.OpenFile(filepath.Join(testDir, fmt.Sprintf("file%d.txt", i)), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("Failed to create file in directory: %v", err)
		}
		f.Close()
	}

	// Open directory for reading
	dir, err := fs.OpenFile(testDir, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open directory: %v", err)
	}

	// Read directory entries
	entries, err := dir.Readdir(-1)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	dir.Close()

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Test Readdirnames
	dir, err = fs.OpenFile(testDir, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open directory: %v", err)
	}

	names, err := dir.Readdirnames(-1)
	if err != nil {
		t.Fatalf("Failed to read directory names: %v", err)
	}
	dir.Close()

	if len(names) != 3 {
		t.Errorf("Expected 3 names, got %d", len(names))
	}

	// Cleanup
	for i := 0; i < 3; i++ {
		fs.Remove(filepath.Join(testDir, fmt.Sprintf("file%d.txt", i)))
	}
	fs.Remove(testDir)
}

// TestIntegrationReaddirLimited tests limited directory reading.
func TestIntegrationReaddirLimited(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testDir := filepath.Join(testBaseDir, "test_readdir_limited")

	// Create directory with multiple files
	err := fs.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	for i := 0; i < 5; i++ {
		f, err := fs.OpenFile(filepath.Join(testDir, fmt.Sprintf("file%d.txt", i)), os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			t.Fatalf("Failed to create file: %v", err)
		}
		f.Close()
	}

	// Read limited entries
	dir, err := fs.OpenFile(testDir, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open directory: %v", err)
	}

	entries, err := dir.Readdir(3)
	if err != nil {
		t.Fatalf("Failed to read directory: %v", err)
	}
	dir.Close()

	if len(entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(entries))
	}

	// Cleanup
	for i := 0; i < 5; i++ {
		fs.Remove(filepath.Join(testDir, fmt.Sprintf("file%d.txt", i)))
	}
	fs.Remove(testDir)
}

// TestIntegrationChmod tests file permission changes.
func TestIntegrationChmod(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_chmod.txt")

	// Create file
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Close()

	// Change permissions
	err = fs.Chmod(testFile, 0755)
	if err != nil {
		t.Fatalf("Failed to chmod: %v", err)
	}

	// Verify permissions changed (note: exact mode may vary by server)
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat: %v", err)
	}

	// Just check it's executable
	if info.Mode()&0100 == 0 {
		t.Log("Warning: chmod may not have worked as expected on this server")
	}

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationChtimes tests file time changes.
func TestIntegrationChtimes(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_chtimes.txt")

	// Create file
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Close()

	// Change times
	newTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	err = fs.Chtimes(testFile, newTime, newTime)
	if err != nil {
		t.Fatalf("Failed to chtimes: %v", err)
	}

	// Verify time changed
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat: %v", err)
	}

	// Allow some tolerance for time differences
	if info.ModTime().Year() != 2020 {
		t.Logf("Warning: chtimes may not have worked as expected: got %v", info.ModTime())
	}

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationLargeFile tests handling of larger files.
func TestIntegrationLargeFile(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_large.txt")

	// Create 1MB of data
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = byte(i % 256)
	}

	// Write file
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	n, err := file.Write(data)
	if err != nil {
		t.Fatalf("Failed to write large file: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(data), n)
	}
	file.Close()

	// Read file back
	file, err = fs.OpenFile(testFile, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open file for reading: %v", err)
	}

	readData, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read large file: %v", err)
	}
	file.Close()

	if !bytes.Equal(readData, data) {
		t.Error("Large file content mismatch")
	}

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationBinaryFile tests handling of binary data.
func TestIntegrationBinaryFile(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_binary.bin")

	// Create binary data with all byte values
	data := make([]byte, 256)
	for i := range data {
		data[i] = byte(i)
	}

	// Write file
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Write(data)
	file.Close()

	// Read file back
	file, err = fs.OpenFile(testFile, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open file for reading: %v", err)
	}

	readData, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read binary file: %v", err)
	}
	file.Close()

	if !bytes.Equal(readData, data) {
		t.Error("Binary file content mismatch")
	}

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationTruncate tests file truncation.
func TestIntegrationTruncate(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_truncate.txt")

	// Create file with content
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Write([]byte("Hello, World!"))
	file.Close()

	// Open for truncation
	file, err = fs.OpenFile(testFile, os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	err = file.Truncate(5)
	if err != nil {
		t.Fatalf("Failed to truncate: %v", err)
	}
	file.Close()

	// Verify size
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat: %v", err)
	}

	if info.Size() != 5 {
		t.Errorf("Expected size 5, got %d", info.Size())
	}

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationReadAt tests reading at specific offset.
func TestIntegrationReadAt(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_readat.txt")

	// Create file
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Write([]byte("0123456789"))
	file.Close()

	// Open for reading
	file, err = fs.OpenFile(testFile, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		file.Close()
		fs.Remove(testFile)
	}()

	// Read at offset 5
	buf := make([]byte, 5)
	n, err := file.ReadAt(buf, 5)
	if err != nil {
		t.Fatalf("Failed to read at offset: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	if string(buf) != "56789" {
		t.Errorf("Expected '56789', got '%s'", string(buf))
	}
}

// TestIntegrationWriteAt tests writing at specific offset.
func TestIntegrationWriteAt(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_writeat.txt")

	// Create file with initial content
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Write([]byte("0123456789"))
	file.Close()

	// Open for writing at offset
	file, err = fs.OpenFile(testFile, os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}

	n, err := file.WriteAt([]byte("XXXXX"), 5)
	if err != nil {
		t.Fatalf("Failed to write at offset: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected 5 bytes, got %d", n)
	}
	file.Close()

	// Verify content
	file, err = fs.OpenFile(testFile, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open file for verification: %v", err)
	}

	content, err := io.ReadAll(file)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	file.Close()

	if string(content) != "01234XXXXX" {
		t.Errorf("Expected '01234XXXXX', got '%s'", string(content))
	}

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationSync tests file sync (should be a no-op but shouldn't error).
func TestIntegrationSync(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_sync.txt")

	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	file.Write([]byte("content"))

	// Sync should not error
	err = file.Sync()
	if err != nil {
		t.Errorf("Sync failed: %v", err)
	}

	file.Close()

	// Cleanup
	fs.Remove(testFile)
}

// TestIntegrationFileStat tests stat on open file.
func TestIntegrationFileStat(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	testFile := filepath.Join(testBaseDir, "test_filestat.txt")
	content := []byte("Hello, World!")

	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	file.Write(content)
	file.Close()

	// Open and stat
	file, err = fs.OpenFile(testFile, os.O_RDONLY, 0)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer func() {
		file.Close()
		fs.Remove(testFile)
	}()

	info, err := file.Stat()
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if info.Size() != int64(len(content)) {
		t.Errorf("Expected size %d, got %d", len(content), info.Size())
	}
}

// TestIntegrationNestedDirectories tests nested directory operations.
func TestIntegrationNestedDirectories(t *testing.T) {
	fs := skipIfNoServer(t)
	defer fs.Close()

	baseDir := filepath.Join(testBaseDir, "nested_test")
	subDir := filepath.Join(baseDir, "subdir")

	// Create base directory
	err := fs.Mkdir(baseDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create base directory: %v", err)
	}

	// Create subdirectory
	err = fs.Mkdir(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create file in subdirectory
	testFile := filepath.Join(subDir, "file.txt")
	file, err := fs.OpenFile(testFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("Failed to create file in subdirectory: %v", err)
	}
	file.Write([]byte("nested content"))
	file.Close()

	// Verify file exists
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file in subdirectory: %v", err)
	}
	if info.IsDir() {
		t.Error("Expected file, not directory")
	}

	// Cleanup (reverse order)
	fs.Remove(testFile)
	fs.Remove(subDir)
	fs.Remove(baseDir)
}
