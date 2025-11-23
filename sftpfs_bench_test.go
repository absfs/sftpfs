package sftpfs

import (
	"os"
	"testing"
)

// Benchmarks for File operations
func BenchmarkFileReaddir(b *testing.B) {
	f := &File{
		dirEntries: make([]os.FileInfo, 100),
		readdirPos: 0,
	}

	// Populate with mock entries
	for i := 0; i < 100; i++ {
		f.dirEntries[i] = &mockFileInfo{name: "file.txt"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.readdirPos = 0 // Reset position
		f.Readdir(-1)
	}
}

func BenchmarkFileReaddirChunked(b *testing.B) {
	f := &File{
		dirEntries: make([]os.FileInfo, 1000),
		readdirPos: 0,
	}

	// Populate with mock entries
	for i := 0; i < 1000; i++ {
		f.dirEntries[i] = &mockFileInfo{name: "file.txt"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.readdirPos = 0 // Reset position
		for {
			entries, err := f.Readdir(10)
			if err != nil || len(entries) == 0 {
				break
			}
		}
	}
}

func BenchmarkFileReaddirnames(b *testing.B) {
	f := &File{
		dirEntries: make([]os.FileInfo, 100),
		readdirPos: 0,
	}

	// Populate with mock entries
	for i := 0; i < 100; i++ {
		f.dirEntries[i] = &mockFileInfo{name: "file.txt"}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.readdirPos = 0 // Reset position
		f.Readdirnames(-1)
	}
}

func BenchmarkWrapError(b *testing.B) {
	err := os.ErrNotExist
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapError("Open", "/path/to/file.txt", err)
	}
}

func BenchmarkWrapErrorf(b *testing.B) {
	err := os.ErrNotExist
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = wrapErrorf("operation failed for %s: %w", "/path/to/file.txt", err)
	}
}

// Benchmark FileSystem methods (these won't actually connect to a server)
func BenchmarkSeparator(b *testing.B) {
	fs := &FileSystem{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.Separator()
	}
}

func BenchmarkListSeparator(b *testing.B) {
	fs := &FileSystem{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.ListSeparator()
	}
}

func BenchmarkGetwd(b *testing.B) {
	fs := &FileSystem{cwd: "/home/user"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = fs.Getwd()
	}
}

func BenchmarkTempDir(b *testing.B) {
	fs := &FileSystem{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fs.TempDir()
	}
}
