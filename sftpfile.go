package sftpfs

import (
	iofs "io/fs"
	"os"
)

// File wraps an sftp.File to implement absfs.File interface.
type File struct {
	file   sftpFileInterface
	name   string
	client sftpClientInterface
}

// Name returns the name of the file.
func (f *File) Name() string {
	return f.name
}

// Read reads from the SFTP file.
func (f *File) Read(b []byte) (int, error) {
	return f.file.Read(b)
}

// ReadAt reads from the SFTP file at a specific offset.
func (f *File) ReadAt(b []byte, off int64) (int, error) {
	return f.file.ReadAt(b, off)
}

// Write writes to the SFTP file.
func (f *File) Write(b []byte) (int, error) {
	return f.file.Write(b)
}

// WriteAt writes to the SFTP file at a specific offset.
func (f *File) WriteAt(b []byte, off int64) (int, error) {
	return f.file.WriteAt(b, off)
}

// WriteString writes a string to the SFTP file.
func (f *File) WriteString(s string) (int, error) {
	return f.file.Write([]byte(s))
}

// Close closes the SFTP file.
func (f *File) Close() error {
	return f.file.Close()
}

// Seek seeks within the SFTP file.
func (f *File) Seek(offset int64, whence int) (int64, error) {
	return f.file.Seek(offset, whence)
}

// Stat returns file info for the SFTP file.
func (f *File) Stat() (os.FileInfo, error) {
	return f.file.Stat()
}

// Sync commits the current contents of the file to stable storage.
func (f *File) Sync() error {
	// SFTP doesn't have a direct sync operation, but we can implement it as a no-op
	// since writes are typically synchronous over the network
	return nil
}

// Truncate changes the size of the file.
func (f *File) Truncate(size int64) error {
	return f.file.Truncate(size)
}

// Readdir reads directory entries.
func (f *File) Readdir(n int) ([]os.FileInfo, error) {
	// Use the client's ReadDir to get directory entries
	entries, err := f.client.ReadDir(f.name)
	if err != nil {
		return nil, err
	}

	// If n <= 0, return all entries
	if n <= 0 {
		return entries, nil
	}

	// Otherwise return up to n entries
	if n > len(entries) {
		n = len(entries)
	}
	return entries[:n], nil
}

// Readdirnames reads directory entry names.
func (f *File) Readdirnames(n int) ([]string, error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}

	names := make([]string, len(infos))
	for i, info := range infos {
		names[i] = info.Name()
	}
	return names, nil
}

// ReadDir reads the directory and returns fs.DirEntry values.
func (f *File) ReadDir(n int) ([]iofs.DirEntry, error) {
	infos, err := f.Readdir(n)
	if err != nil {
		return nil, err
	}

	entries := make([]iofs.DirEntry, len(infos))
	for i, info := range infos {
		entries[i] = &dirEntry{info: info}
	}
	return entries, nil
}
