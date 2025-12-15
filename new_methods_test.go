package sftpfs

import (
	"os"
	"testing"

	"github.com/absfs/sftpfs/internal/mocks"
)

func TestNewMethods(t *testing.T) {
	// Create a raw FileSystem for testing, not the wrapped version
	client := newEnhancedMockSFTPClient()
	sshClient := &mocks.MockSSHClient{}
	client.dirs["/"] = []os.FileInfo{}
	client.dirs["/tmp"] = []os.FileInfo{}
	client.permissions["/"] = os.ModeDir | 0755
	client.permissions["/tmp"] = os.ModeDir | 0755

	fs := newWithClients(client, sshClient)

	// Test ReadDir
	err := fs.Mkdir("/test_readdir", 0755)
	if err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	// Create some files in the directory
	f1, _ := fs.OpenFile("/test_readdir/file1.txt", os.O_CREATE|os.O_RDWR, 0644)
	f1.Close()
	f2, _ := fs.OpenFile("/test_readdir/file2.txt", os.O_CREATE|os.O_RDWR, 0644)
	f2.Close()

	entries, err := fs.ReadDir("/test_readdir")
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(entries))
	}

	// Test ReadFile
	testData := []byte("test content")
	f, _ := fs.OpenFile("/test_readfile.txt", os.O_CREATE|os.O_RDWR, 0644)
	f.Write(testData)
	f.Close()

	data, err := fs.ReadFile("/test_readfile.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}
	if string(data) != string(testData) {
		t.Errorf("Expected %q, got %q", string(testData), string(data))
	}

	// Test Sub
	fs.Mkdir("/sub", 0755)
	fs.Mkdir("/sub/dir", 0755)
	subfs, err := fs.Sub("/sub")
	if err != nil {
		t.Fatalf("Sub failed: %v", err)
	}

	// Create a file in the parent fs first (Sub returns fs.FS which is read-only)
	subf, err := fs.OpenFile("/sub/test.txt", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		t.Fatalf("Create in subfs failed: %v", err)
	}
	subf.Write([]byte("sub test"))
	subf.Close()

	// Verify we can read it through subfs using fs.FS Open method
	subFile, err := subfs.Open("test.txt")
	if err != nil {
		t.Fatalf("Open in subfs failed: %v", err)
	}
	subFile.Close()

	// Verify it exists in the parent fs
	info, err := fs.Stat("/sub/test.txt")
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}
	if info.IsDir() {
		t.Error("Expected file, got directory")
	}
}
