// Package mocks provides test doubles for SFTP testing.
package mocks

import (
	"io"
	"io/fs"
	"os"
	"time"
)

// MockFileInfo implements os.FileInfo for testing.
type MockFileInfo struct {
	FileName    string
	FileSize    int64
	FileMode    os.FileMode
	FileModTime time.Time
	FileIsDir   bool
	FileSys     any
}

func (m *MockFileInfo) Name() string       { return m.FileName }
func (m *MockFileInfo) Size() int64        { return m.FileSize }
func (m *MockFileInfo) Mode() os.FileMode  { return m.FileMode }
func (m *MockFileInfo) ModTime() time.Time { return m.FileModTime }
func (m *MockFileInfo) IsDir() bool        { return m.FileIsDir }
func (m *MockFileInfo) Sys() any           { return m.FileSys }

// MockSFTPFile is a configurable mock for sftp.File.
type MockSFTPFile struct {
	Data        []byte
	Position    int64
	CloseErr    error
	ReadErr     error
	WriteErr    error
	SeekErr     error
	StatErr     error
	TruncateErr error
	StatInfo    os.FileInfo
	Closed      bool
}

func (f *MockSFTPFile) Read(b []byte) (int, error) {
	if f.ReadErr != nil {
		return 0, f.ReadErr
	}
	if f.Position >= int64(len(f.Data)) {
		return 0, io.EOF
	}
	n := copy(b, f.Data[f.Position:])
	f.Position += int64(n)
	return n, nil
}

func (f *MockSFTPFile) ReadAt(b []byte, off int64) (int, error) {
	if f.ReadErr != nil {
		return 0, f.ReadErr
	}
	if off >= int64(len(f.Data)) {
		return 0, io.EOF
	}
	n := copy(b, f.Data[off:])
	if n < len(b) {
		return n, io.EOF
	}
	return n, nil
}

func (f *MockSFTPFile) Write(b []byte) (int, error) {
	if f.WriteErr != nil {
		return 0, f.WriteErr
	}
	// Expand data if necessary
	needed := int(f.Position) + len(b)
	if needed > len(f.Data) {
		newData := make([]byte, needed)
		copy(newData, f.Data)
		f.Data = newData
	}
	n := copy(f.Data[f.Position:], b)
	f.Position += int64(n)
	return n, nil
}

func (f *MockSFTPFile) WriteAt(b []byte, off int64) (int, error) {
	if f.WriteErr != nil {
		return 0, f.WriteErr
	}
	needed := int(off) + len(b)
	if needed > len(f.Data) {
		newData := make([]byte, needed)
		copy(newData, f.Data)
		f.Data = newData
	}
	n := copy(f.Data[off:], b)
	return n, nil
}

func (f *MockSFTPFile) Seek(offset int64, whence int) (int64, error) {
	if f.SeekErr != nil {
		return 0, f.SeekErr
	}
	var newPos int64
	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = f.Position + offset
	case io.SeekEnd:
		newPos = int64(len(f.Data)) + offset
	}
	if newPos < 0 {
		return 0, fs.ErrInvalid
	}
	f.Position = newPos
	return newPos, nil
}

func (f *MockSFTPFile) Close() error {
	f.Closed = true
	return f.CloseErr
}

func (f *MockSFTPFile) Stat() (os.FileInfo, error) {
	if f.StatErr != nil {
		return nil, f.StatErr
	}
	if f.StatInfo != nil {
		return f.StatInfo, nil
	}
	return &MockFileInfo{
		FileSize: int64(len(f.Data)),
		FileMode: 0644,
	}, nil
}

func (f *MockSFTPFile) Truncate(size int64) error {
	if f.TruncateErr != nil {
		return f.TruncateErr
	}
	if size < int64(len(f.Data)) {
		f.Data = f.Data[:size]
	} else if size > int64(len(f.Data)) {
		newData := make([]byte, size)
		copy(newData, f.Data)
		f.Data = newData
	}
	return nil
}

// MockSSHClient is a mock SSH client for testing.
type MockSSHClient struct {
	CloseErr error
	Closed   bool
}

func (c *MockSSHClient) Close() error {
	c.Closed = true
	return c.CloseErr
}
