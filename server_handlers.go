package sftpfs

import (
	"io"
	"os"
	"path"
	"sort"
	"sync"
	"time"

	"github.com/absfs/absfs"
	"github.com/pkg/sftp"
)

// ServerHandler implements all four sftp.Handlers interfaces:
// FileReader, FileWriter, FileCmder, and FileLister.
// It adapts an absfs.FileSystem to serve files via SFTP protocol.
type ServerHandler struct {
	fs absfs.FileSystem
	mu sync.RWMutex
}

// NewServerHandler creates SFTP handlers that serve the given absfs.FileSystem.
func NewServerHandler(fs absfs.FileSystem) sftp.Handlers {
	h := &ServerHandler{fs: fs}
	return sftp.Handlers{
		FileGet:  h,
		FilePut:  h,
		FileCmd:  h,
		FileList: h,
	}
}

// Fileread implements sftp.FileReader.
// Returns an io.ReaderAt for the requested file path.
// Called for SFTP Method: Get
func (h *ServerHandler) Fileread(r *sftp.Request) (io.ReaderAt, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	f, err := h.fs.Open(r.Filepath)
	if err != nil {
		return nil, err
	}
	return &serverFile{file: f, path: r.Filepath}, nil
}

// Filewrite implements sftp.FileWriter.
// Returns an io.WriterAt for the requested file path.
// Called for SFTP Methods: Put, Open
func (h *ServerHandler) Filewrite(r *sftp.Request) (io.WriterAt, error) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Determine flags from the request
	flags := os.O_WRONLY | os.O_CREATE

	// Check pflags for append/truncate behavior
	pflags := r.Pflags()
	if pflags.Append {
		flags |= os.O_APPEND
	}
	if pflags.Trunc {
		flags |= os.O_TRUNC
	}
	if pflags.Excl {
		flags |= os.O_EXCL
	}
	if pflags.Read && pflags.Write {
		flags = os.O_RDWR | os.O_CREATE
	}

	f, err := h.fs.OpenFile(r.Filepath, flags, 0644)
	if err != nil {
		return nil, err
	}
	return &serverFile{file: f, path: r.Filepath}, nil
}

// Filecmd implements sftp.FileCmder.
// Handles file commands like mkdir, remove, rename, etc.
// Called for SFTP Methods: Setstat, Rename, Rmdir, Mkdir, Link, Symlink, Remove
func (h *ServerHandler) Filecmd(r *sftp.Request) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	switch r.Method {
	case "Setstat":
		return h.handleSetstat(r)
	case "Rename":
		return h.fs.Rename(r.Filepath, r.Target)
	case "Rmdir":
		return h.fs.Remove(r.Filepath)
	case "Mkdir":
		return h.fs.Mkdir(r.Filepath, 0755)
	case "Remove":
		return h.fs.Remove(r.Filepath)
	case "Symlink":
		if sfs, ok := h.fs.(absfs.SymlinkFileSystem); ok {
			return sfs.Symlink(r.Target, r.Filepath)
		}
		return sftp.ErrSSHFxOpUnsupported
	case "Link":
		// Hard links not commonly supported
		return sftp.ErrSSHFxOpUnsupported
	default:
		return sftp.ErrSSHFxOpUnsupported
	}
}

// handleSetstat handles the Setstat command for changing file attributes.
func (h *ServerHandler) handleSetstat(r *sftp.Request) error {
	attrs := r.Attributes()

	// Handle mode change
	if attrs.FileMode() != 0 {
		if err := h.fs.Chmod(r.Filepath, attrs.FileMode()); err != nil {
			return err
		}
	}

	// Handle time changes (Atime and Mtime are uint32 Unix timestamps)
	if attrs.Atime != 0 || attrs.Mtime != 0 {
		atime := time.Unix(int64(attrs.Atime), 0)
		mtime := time.Unix(int64(attrs.Mtime), 0)
		if attrs.Atime == 0 {
			atime = mtime
		}
		if attrs.Mtime == 0 {
			mtime = atime
		}
		if err := h.fs.Chtimes(r.Filepath, atime, mtime); err != nil {
			return err
		}
	}

	// Handle ownership changes
	if attrs.UID != 0 || attrs.GID != 0 {
		if err := h.fs.Chown(r.Filepath, int(attrs.UID), int(attrs.GID)); err != nil {
			return err
		}
	}

	return nil
}

// Filelist implements sftp.FileLister.
// Returns a ListerAt for directory listings and file stat operations.
// Called for SFTP Methods: List, Stat, Readlink
func (h *ServerHandler) Filelist(r *sftp.Request) (sftp.ListerAt, error) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch r.Method {
	case "List":
		return h.handleList(r)
	case "Stat":
		return h.handleStat(r)
	case "Readlink":
		return h.handleReadlink(r)
	default:
		return nil, sftp.ErrSSHFxOpUnsupported
	}
}

// handleList returns directory contents.
func (h *ServerHandler) handleList(r *sftp.Request) (sftp.ListerAt, error) {
	dir, err := h.fs.Open(r.Filepath)
	if err != nil {
		return nil, err
	}
	defer dir.Close()

	entries, err := dir.Readdir(-1)
	if err != nil {
		return nil, err
	}

	// Sort entries by name for consistent ordering
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	return &listerat{entries: entries}, nil
}

// handleStat returns file info for a single file.
func (h *ServerHandler) handleStat(r *sftp.Request) (sftp.ListerAt, error) {
	info, err := h.fs.Stat(r.Filepath)
	if err != nil {
		return nil, err
	}
	return &listerat{entries: []os.FileInfo{info}}, nil
}

// handleReadlink returns the target of a symbolic link.
func (h *ServerHandler) handleReadlink(r *sftp.Request) (sftp.ListerAt, error) {
	sfs, ok := h.fs.(absfs.SymlinkFileSystem)
	if !ok {
		return nil, sftp.ErrSSHFxOpUnsupported
	}

	target, err := sfs.Readlink(r.Filepath)
	if err != nil {
		return nil, err
	}

	// Return a fake FileInfo with the link target as the name
	return &listerat{entries: []os.FileInfo{&linkInfo{name: target}}}, nil
}

// serverFile wraps an absfs.File to implement io.ReaderAt, io.WriterAt, and io.Closer.
type serverFile struct {
	file absfs.File
	path string
	mu   sync.Mutex
}

// ReadAt implements io.ReaderAt.
func (f *serverFile) ReadAt(p []byte, off int64) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, err := f.file.Seek(off, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return f.file.Read(p)
}

// WriteAt implements io.WriterAt.
func (f *serverFile) WriteAt(p []byte, off int64) (int, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	_, err := f.file.Seek(off, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return f.file.Write(p)
}

// Close implements io.Closer.
func (f *serverFile) Close() error {
	return f.file.Close()
}

// listerat implements sftp.ListerAt for directory listings.
type listerat struct {
	entries []os.FileInfo
}

// ListAt implements sftp.ListerAt.
// Copies FileInfo entries into the provided buffer starting at offset.
func (l *listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	if offset >= int64(len(l.entries)) {
		return 0, io.EOF
	}

	n := copy(ls, l.entries[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}

// linkInfo is a minimal FileInfo for symlink targets.
type linkInfo struct {
	name string
}

func (l *linkInfo) Name() string       { return path.Base(l.name) }
func (l *linkInfo) Size() int64        { return int64(len(l.name)) }
func (l *linkInfo) Mode() os.FileMode  { return os.ModeSymlink | 0777 }
func (l *linkInfo) ModTime() time.Time { return time.Time{} }
func (l *linkInfo) IsDir() bool        { return false }
func (l *linkInfo) Sys() interface{}   { return nil }
