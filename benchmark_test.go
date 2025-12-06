package sftpfs

import (
	"os"
	"testing"

	"github.com/absfs/sftpfs/internal/mocks"
)

// BenchmarkFileRead benchmarks reading from a file.
func BenchmarkFileRead(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	mockFile := &mocks.MockSFTPFile{Data: data}
	file := &File{file: mockFile, name: "/test.txt"}

	buf := make([]byte, 64)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockFile.Position = 0
		file.Read(buf)
	}
}

// BenchmarkFileWrite benchmarks writing to a file.
func BenchmarkFileWrite(b *testing.B) {
	mockFile := &mocks.MockSFTPFile{Data: make([]byte, 0, 4096)}
	file := &File{file: mockFile, name: "/test.txt"}

	data := make([]byte, 64)
	for i := range data {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockFile.Position = 0
		mockFile.Data = mockFile.Data[:0]
		file.Write(data)
	}
}

// BenchmarkFileSeek benchmarks seeking within a file.
func BenchmarkFileSeek(b *testing.B) {
	mockFile := &mocks.MockSFTPFile{Data: make([]byte, 1024)}
	file := &File{file: mockFile, name: "/test.txt"}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		file.Seek(int64(i%1024), 0)
	}
}

// BenchmarkReaddir benchmarks directory listing.
func BenchmarkReaddir(b *testing.B) {
	mockClient := newMockSFTPClient()
	entries := make([]os.FileInfo, 100)
	for i := range entries {
		entries[i] = &mocks.MockFileInfo{FileName: "file" + string(rune('0'+i%10))}
	}
	mockClient.dirs["/testdir"] = entries

	file := &File{
		file:   &mocks.MockSFTPFile{},
		name:   "/testdir",
		client: mockClient,
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		file.Readdir(-1)
	}
}

// BenchmarkStat benchmarks file stat operations.
func BenchmarkStat(b *testing.B) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{Data: []byte("hello world")}
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fs.Stat("/test.txt")
	}
}

// BenchmarkOpenFile benchmarks file open operations.
func BenchmarkOpenFile(b *testing.B) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{Data: []byte("hello world")}
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		f, _ := fs.OpenFile("/test.txt", os.O_RDONLY, 0644)
		f.Close()
	}
}

// BenchmarkMkdir benchmarks directory creation.
func BenchmarkMkdir(b *testing.B) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Reset dirs for each iteration
		mockClient.dirs = make(map[string][]os.FileInfo)
		fs.Mkdir("/testdir", 0755)
	}
}

// BenchmarkRename benchmarks file rename operations.
func BenchmarkRename(b *testing.B) {
	mockClient := newMockSFTPClient()
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		mockClient.files["/old.txt"] = &mocks.MockSFTPFile{Data: []byte("content")}
		fs.Rename("/old.txt", "/new.txt")
	}
}

// BenchmarkChmod benchmarks chmod operations.
func BenchmarkChmod(b *testing.B) {
	mockClient := newMockSFTPClient()
	mockClient.files["/test.txt"] = &mocks.MockSFTPFile{}
	fs := newWithClients(mockClient, &mocks.MockSSHClient{})

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fs.Chmod("/test.txt", 0755)
	}
}
