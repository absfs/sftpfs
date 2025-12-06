package sftpfs

import (
	"io"
	"os"
	"testing"
	"time"

	"github.com/absfs/absfs"
	"github.com/absfs/sftpfs/internal/mocks"
)

// TestFileSystemImplementsFiler verifies that FileSystem implements absfs.Filer.
func TestFileSystemImplementsFiler(t *testing.T) {
	var _ absfs.Filer = (*FileSystem)(nil)
}

// TestFileImplementsAbsfsFile verifies that File implements absfs.File.
func TestFileImplementsAbsfsFile(t *testing.T) {
	var _ absfs.File = (*File)(nil)
}

// TestFileImplementsInterfaces verifies File implements standard interfaces.
func TestFileImplementsInterfaces(t *testing.T) {
	// io.Reader
	var _ io.Reader = (*File)(nil)

	// io.Writer
	var _ io.Writer = (*File)(nil)

	// io.Seeker
	var _ io.Seeker = (*File)(nil)

	// io.Closer
	var _ io.Closer = (*File)(nil)

	// io.ReaderAt
	var _ io.ReaderAt = (*File)(nil)

	// io.WriterAt
	var _ io.WriterAt = (*File)(nil)
}

// TestFilerMethodsExist verifies all Filer interface methods exist.
func TestFilerMethodsExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{Data: []byte("test")}
	mockClient.dirs["/testdir"] = []os.FileInfo{}

	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	// OpenFile
	f, err := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
	if err != nil {
		t.Errorf("OpenFile failed: %v", err)
	} else {
		f.Close()
	}

	// Mkdir
	err = fs.Mkdir("/newdir", 0755)
	if err != nil {
		t.Errorf("Mkdir failed: %v", err)
	}

	// Remove
	err = fs.Remove("/test.txt")
	if err != nil {
		t.Errorf("Remove failed: %v", err)
	}

	// Rename
	mockClient.files["/old.txt"] = &mocks.MockSFTPFile{}
	err = fs.Rename("/old.txt", "/new.txt")
	if err != nil {
		t.Errorf("Rename failed: %v", err)
	}

	// Stat
	_, err = fs.Stat("/testdir")
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}

	// Chmod
	mockClient.files["/file.txt"] = &mocks.MockSFTPFile{}
	err = fs.Chmod("/file.txt", 0755)
	if err != nil {
		t.Errorf("Chmod failed: %v", err)
	}

	// Chtimes
	now := time.Now()
	err = fs.Chtimes("/file.txt", now, now)
	if err != nil {
		t.Errorf("Chtimes failed: %v", err)
	}

	// Chown
	err = fs.Chown("/file.txt", 1000, 1000)
	if err != nil {
		t.Errorf("Chown failed: %v", err)
	}
}

// TestFileMethodsExist verifies all File interface methods exist.
func TestFileMethodsExist(t *testing.T) {
	mockClient := newMockSFTPClient()
	mockFile := &mocks.MockSFTPFile{
		Data: []byte("hello world"),
		StatInfo: &mocks.MockFileInfo{
			FileName: "test.txt",
			FileSize: 11,
		},
	}
	mockClient.files["/test.txt"] = mockFile
	mockClient.dirs["/test.txt"] = []os.FileInfo{}

	file := &File{
		file:   mockFile,
		name:   "/test.txt",
		client: mockClient,
	}

	// Name
	name := file.Name()
	if name != "/test.txt" {
		t.Errorf("Name() = %q, want %q", name, "/test.txt")
	}

	// Read
	buf := make([]byte, 5)
	n, err := file.Read(buf)
	if err != nil {
		t.Errorf("Read failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Read returned %d, want 5", n)
	}

	// Write
	mockFile.Position = 0
	n, err = file.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write failed: %v", err)
	}
	if n != 4 {
		t.Errorf("Write returned %d, want 4", n)
	}

	// WriteAt
	n, err = file.WriteAt([]byte("xx"), 0)
	if err != nil {
		t.Errorf("WriteAt failed: %v", err)
	}
	if n != 2 {
		t.Errorf("WriteAt returned %d, want 2", n)
	}

	// WriteString
	n, err = file.WriteString("str")
	if err != nil {
		t.Errorf("WriteString failed: %v", err)
	}
	if n != 3 {
		t.Errorf("WriteString returned %d, want 3", n)
	}

	// ReadAt
	mockFile.Data = []byte("hello world")
	n, err = file.ReadAt(buf, 0)
	if err != nil {
		t.Errorf("ReadAt failed: %v", err)
	}
	if n != 5 {
		t.Errorf("ReadAt returned %d, want 5", n)
	}

	// Seek
	pos, err := file.Seek(0, io.SeekStart)
	if err != nil {
		t.Errorf("Seek failed: %v", err)
	}
	if pos != 0 {
		t.Errorf("Seek returned %d, want 0", pos)
	}

	// Stat
	info, err := file.Stat()
	if err != nil {
		t.Errorf("Stat failed: %v", err)
	}
	if info == nil {
		t.Error("Stat returned nil FileInfo")
	}

	// Sync
	err = file.Sync()
	if err != nil {
		t.Errorf("Sync failed: %v", err)
	}

	// Truncate
	err = file.Truncate(5)
	if err != nil {
		t.Errorf("Truncate failed: %v", err)
	}

	// Readdir
	entries, err := file.Readdir(-1)
	if err != nil {
		t.Errorf("Readdir failed: %v", err)
	}
	if entries == nil {
		t.Error("Readdir returned nil")
	}

	// Readdirnames
	names, err := file.Readdirnames(-1)
	if err != nil {
		t.Errorf("Readdirnames failed: %v", err)
	}
	if names == nil {
		t.Error("Readdirnames returned nil")
	}

	// Close
	err = file.Close()
	if err != nil {
		t.Errorf("Close failed: %v", err)
	}
}

// TestAbsfsFileBehavior tests that our File behaves like other absfs implementations.
func TestAbsfsFileBehavior(t *testing.T) {
	mockClient := newMockSFTPClient()

	t.Run("sequential read/write", func(t *testing.T) {
		mockFile := &mocks.MockSFTPFile{Data: []byte{}}
		mockClient.files["/seqtest.txt"] = mockFile

		file := &File{file: mockFile, name: "/seqtest.txt", client: mockClient}

		// Write some data
		n, err := file.Write([]byte("hello"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != 5 {
			t.Errorf("Write returned %d, want 5", n)
		}

		// Write more data
		n, err = file.Write([]byte(" world"))
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		if n != 6 {
			t.Errorf("Write returned %d, want 6", n)
		}

		// Verify data
		if string(mockFile.Data) != "hello world" {
			t.Errorf("Data = %q, want %q", string(mockFile.Data), "hello world")
		}
	})

	t.Run("seek and read", func(t *testing.T) {
		mockFile := &mocks.MockSFTPFile{Data: []byte("0123456789")}
		file := &File{file: mockFile, name: "/seektest.txt", client: mockClient}

		// Seek to middle
		pos, err := file.Seek(5, io.SeekStart)
		if err != nil {
			t.Fatalf("Seek failed: %v", err)
		}
		if pos != 5 {
			t.Errorf("Seek returned %d, want 5", pos)
		}

		// Read from position
		buf := make([]byte, 5)
		n, err := file.Read(buf)
		if err != nil {
			t.Fatalf("Read failed: %v", err)
		}
		if n != 5 {
			t.Errorf("Read returned %d, want 5", n)
		}
		if string(buf) != "56789" {
			t.Errorf("Read data = %q, want %q", string(buf), "56789")
		}
	})

	t.Run("read at EOF returns io.EOF", func(t *testing.T) {
		mockFile := &mocks.MockSFTPFile{Data: []byte("test")}
		file := &File{file: mockFile, name: "/eoftest.txt", client: mockClient}

		buf := make([]byte, 10)

		// First read should succeed
		n, err := file.Read(buf)
		if err != nil {
			t.Fatalf("First read failed: %v", err)
		}
		if n != 4 {
			t.Errorf("First read returned %d, want 4", n)
		}

		// Second read should return EOF
		n, err = file.Read(buf)
		if err != io.EOF {
			t.Errorf("Second read error = %v, want io.EOF", err)
		}
		if n != 0 {
			t.Errorf("Second read returned %d, want 0", n)
		}
	})

	t.Run("truncate shrinks file", func(t *testing.T) {
		mockFile := &mocks.MockSFTPFile{Data: []byte("hello world")}
		file := &File{file: mockFile, name: "/trunctest.txt", client: mockClient}

		err := file.Truncate(5)
		if err != nil {
			t.Fatalf("Truncate failed: %v", err)
		}

		if len(mockFile.Data) != 5 {
			t.Errorf("Data length = %d, want 5", len(mockFile.Data))
		}
		if string(mockFile.Data) != "hello" {
			t.Errorf("Data = %q, want %q", string(mockFile.Data), "hello")
		}
	})

	t.Run("truncate expands file", func(t *testing.T) {
		mockFile := &mocks.MockSFTPFile{Data: []byte("hi")}
		file := &File{file: mockFile, name: "/trunctest2.txt", client: mockClient}

		err := file.Truncate(10)
		if err != nil {
			t.Fatalf("Truncate failed: %v", err)
		}

		if len(mockFile.Data) != 10 {
			t.Errorf("Data length = %d, want 10", len(mockFile.Data))
		}
	})
}

// TestFilerBehavior tests filesystem-level behavior.
func TestFilerBehavior(t *testing.T) {
	t.Run("mkdir creates directory", func(t *testing.T) {
		mockClient := newMockSFTPClient()
		fs := newWithClients(mockClient, &mocks.MockSSHClient{})

		err := fs.Mkdir("/mydir", 0755)
		if err != nil {
			t.Fatalf("Mkdir failed: %v", err)
		}

		if _, exists := mockClient.dirs["/mydir"]; !exists {
			t.Error("Directory was not created")
		}
	})

	t.Run("remove deletes file", func(t *testing.T) {
		mockClient := newMockSFTPClient()
		mockClient.files["/myfile.txt"] = &mocks.MockSFTPFile{}
		fs := newWithClients(mockClient, &mocks.MockSSHClient{})

		err := fs.Remove("/myfile.txt")
		if err != nil {
			t.Fatalf("Remove failed: %v", err)
		}

		if _, exists := mockClient.files["/myfile.txt"]; exists {
			t.Error("File was not deleted")
		}
	})

	t.Run("rename moves file", func(t *testing.T) {
		mockClient := newMockSFTPClient()
		mockClient.files["/old.txt"] = &mocks.MockSFTPFile{Data: []byte("content")}
		fs := newWithClients(mockClient, &mocks.MockSSHClient{})

		err := fs.Rename("/old.txt", "/new.txt")
		if err != nil {
			t.Fatalf("Rename failed: %v", err)
		}

		if _, exists := mockClient.files["/old.txt"]; exists {
			t.Error("Old file still exists")
		}
		if _, exists := mockClient.files["/new.txt"]; !exists {
			t.Error("New file does not exist")
		}
	})

	t.Run("stat returns correct info", func(t *testing.T) {
		mockClient := newMockSFTPClient()
		mockClient.files["/test.txt"] = &mocks.MockSFTPFile{Data: []byte("hello world")}
		fs := newWithClients(mockClient, &mocks.MockSSHClient{})

		info, err := fs.Stat("/test.txt")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}

		if info.Size() != 11 {
			t.Errorf("Size = %d, want 11", info.Size())
		}
	})

	t.Run("stat directory returns IsDir true", func(t *testing.T) {
		mockClient := newMockSFTPClient()
		mockClient.dirs["/mydir"] = []os.FileInfo{}
		fs := newWithClients(mockClient, &mocks.MockSSHClient{})

		info, err := fs.Stat("/mydir")
		if err != nil {
			t.Fatalf("Stat failed: %v", err)
		}

		if !info.IsDir() {
			t.Error("IsDir() = false, want true")
		}
	})
}
