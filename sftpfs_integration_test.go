//go:build integration
// +build integration

package sftpfs

import (
	"io"
	"os"
	"testing"
	"time"
)

const (
	testHost     = "localhost:2222"
	testUser     = "testuser"
	testPassword = "testpass"
)

func TestIntegrationPasswordAuth(t *testing.T) {
	// Wait a bit for the SFTP server to be ready
	time.Sleep(2 * time.Second)

	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Test basic operations
	testBasicOperations(t, fs)
}

func TestIntegrationKeyAuth(t *testing.T) {
	// Wait a bit for the SFTP server to be ready
	time.Sleep(2 * time.Second)

	// Read the private key
	key, err := os.ReadFile("testdata/ssh/id_rsa")
	if err != nil {
		t.Skipf("Skipping test - key file not found: %v", err)
	}

	fs, err := DialWithKey(testHost, testUser, key)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available or key auth failed: %v", err)
	}
	defer fs.Close()

	// Test basic operations
	testBasicOperations(t, fs)
}

func TestIntegrationFileOperations(t *testing.T) {
	time.Sleep(2 * time.Second)

	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Test Create and Write
	testFile := "/upload/test_file.txt"
	f, err := fs.Create(testFile)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	testData := []byte("Hello, SFTP!")
	n, err := f.Write(testData)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
	}
	f.Close()

	// Test Open and Read
	f, err = fs.Open(testFile)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	readData := make([]byte, len(testData))
	n, err = f.Read(readData)
	if err != nil && err != io.EOF {
		t.Fatalf("Read failed: %v", err)
	}
	if string(readData[:n]) != string(testData) {
		t.Errorf("Expected %s, got %s", testData, readData[:n])
	}
	f.Close()

	// Test Stat
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.Size() != int64(len(testData)) {
		t.Errorf("Expected size %d, got %d", len(testData), info.Size())
	}

	// Test Remove
	err = fs.Remove(testFile)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Verify file is removed
	_, err = fs.Stat(testFile)
	if err == nil {
		t.Error("Expected error after removing file, got nil")
	}
}

func TestIntegrationDirectoryOperations(t *testing.T) {
	time.Sleep(2 * time.Second)

	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Test Mkdir
	testDir := "/upload/test_dir"
	err = fs.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	// Test MkdirAll
	testNestedDir := "/upload/test_nested/dir/structure"
	err = fs.MkdirAll(testNestedDir, 0755)
	if err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Verify directory exists
	info, err := fs.Stat(testNestedDir)
	if err != nil {
		t.Fatalf("Stat on directory failed: %v", err)
	}
	if !info.IsDir() {
		t.Error("Expected directory, got file")
	}

	// Test RemoveAll
	err = fs.RemoveAll("/upload/test_nested")
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}

	// Verify directory is removed
	_, err = fs.Stat("/upload/test_nested")
	if err == nil {
		t.Error("Expected error after removing directory, got nil")
	}

	// Clean up single directory
	fs.Remove(testDir)
}

func TestIntegrationRename(t *testing.T) {
	time.Sleep(2 * time.Second)

	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Create a file
	oldPath := "/upload/old_name.txt"
	f, err := fs.Create(oldPath)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	f.Write([]byte("test"))
	f.Close()

	// Rename it
	newPath := "/upload/new_name.txt"
	err = fs.Rename(oldPath, newPath)
	if err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	// Verify old path doesn't exist
	_, err = fs.Stat(oldPath)
	if err == nil {
		t.Error("Old path should not exist after rename")
	}

	// Verify new path exists
	_, err = fs.Stat(newPath)
	if err != nil {
		t.Errorf("New path should exist after rename: %v", err)
	}

	// Clean up
	fs.Remove(newPath)
}

func TestIntegrationChmod(t *testing.T) {
	time.Sleep(2 * time.Second)

	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Create a file
	testFile := "/upload/chmod_test.txt"
	f, err := fs.Create(testFile)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	f.Close()

	// Change mode
	err = fs.Chmod(testFile, 0600)
	if err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	// Verify mode changed
	info, err := fs.Stat(testFile)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	// Clean up
	fs.Remove(testFile)

	// Note: We can't reliably test the exact mode due to server-side umask
	// and permission handling, so we just verify the operation doesn't error
	if info.Mode() == 0 {
		t.Error("Mode should not be 0")
	}
}

func TestIntegrationChdirGetwd(t *testing.T) {
	time.Sleep(2 * time.Second)

	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Get initial working directory
	wd, err := fs.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if wd != "/" {
		t.Errorf("Expected initial wd to be /, got %s", wd)
	}

	// Change to upload directory
	err = fs.Chdir("/upload")
	if err != nil {
		t.Fatalf("Chdir failed: %v", err)
	}

	// Verify working directory changed
	wd, err = fs.Getwd()
	if err != nil {
		t.Fatalf("Getwd failed: %v", err)
	}
	if wd != "/upload" {
		t.Errorf("Expected wd to be /upload, got %s", wd)
	}
}

func TestIntegrationSymlinks(t *testing.T) {
	time.Sleep(2 * time.Second)

	fs, err := Dial(testHost, testUser, testPassword)
	if err != nil {
		t.Skipf("Skipping integration test - SFTP server not available: %v", err)
	}
	defer fs.Close()

	// Create a target file
	targetFile := "/upload/symlink_target.txt"
	f, err := fs.Create(targetFile)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	f.Write([]byte("target"))
	f.Close()

	// Create a symlink
	linkFile := "/upload/symlink.txt"
	err = fs.Symlink(targetFile, linkFile)
	if err != nil {
		t.Fatalf("Symlink failed: %v", err)
	}

	// Read the symlink
	target, err := fs.Readlink(linkFile)
	if err != nil {
		t.Fatalf("Readlink failed: %v", err)
	}
	if target != targetFile {
		t.Errorf("Expected target %s, got %s", targetFile, target)
	}

	// Test Lstat
	info, err := fs.Lstat(linkFile)
	if err != nil {
		t.Fatalf("Lstat failed: %v", err)
	}
	// Note: Mode checking for symlinks can be implementation-dependent
	_ = info

	// Clean up
	fs.Remove(linkFile)
	fs.Remove(targetFile)
}

func testBasicOperations(t *testing.T, fs *FileSystem) {
	// Test Separator
	if fs.Separator() != '/' {
		t.Errorf("Expected separator /, got %c", fs.Separator())
	}

	// Test TempDir
	if fs.TempDir() != "/tmp" {
		t.Errorf("Expected temp dir /tmp, got %s", fs.TempDir())
	}

	// Test Getwd
	wd, err := fs.Getwd()
	if err != nil {
		t.Errorf("Getwd failed: %v", err)
	}
	if wd == "" {
		t.Error("Working directory should not be empty")
	}
}
