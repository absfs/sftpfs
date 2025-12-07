# SftpFs - SFTP FileSystem

[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/sftpfs.svg)](https://pkg.go.dev/github.com/absfs/sftpfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/absfs/sftpfs)](https://goreportcard.com/report/github.com/absfs/sftpfs)
[![CI](https://github.com/absfs/sftpfs/actions/workflows/ci.yml/badge.svg)](https://github.com/absfs/sftpfs/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

The `sftpfs` package implements an `absfs.Filer` for SFTP (SSH File Transfer Protocol). It provides secure file operations over SSH using the `github.com/pkg/sftp` library.

## Features

- **Secure file operations**: All operations encrypted over SSH
- **Multiple authentication methods**: Password and SSH key authentication
- **Standard interface**: Implements `absfs.Filer` for seamless integration
- **Full file operations**: Read, write, seek, truncate, and more
- **Directory operations**: Create, remove, and list directories

## Install

```bash
go get github.com/absfs/sftpfs
```

## Example Usage

### Password Authentication

```go
package main

import (
    "log"
    "os"

    "github.com/absfs/sftpfs"
)

func main() {
    // Connect using password
    fs, err := sftpfs.Dial("example.com:22", "username", "password")
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Use like any other filesystem
    f, _ := fs.OpenFile("/remote/path/file.txt", os.O_RDONLY, 0)
    defer f.Close()

    // Read, write, etc.
}
```

### Key-Based Authentication

```go
package main

import (
    "log"
    "os"

    "github.com/absfs/sftpfs"
)

func main() {
    // Read private key
    key, err := os.ReadFile("/home/user/.ssh/id_rsa")
    if err != nil {
        log.Fatal(err)
    }

    // Connect using SSH key
    fs, err := sftpfs.DialWithKey("example.com:22", "username", key)
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Use filesystem operations
    fs.Mkdir("/remote/newdir", 0755)
}
```

### Advanced Configuration

```go
package main

import (
    "log"
    "time"

    "github.com/absfs/sftpfs"
)

func main() {
    config := &sftpfs.Config{
        Host:     "example.com:22",
        User:     "username",
        Password: "password",
        Timeout:  60 * time.Second,
    }

    fs, err := sftpfs.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Use filesystem
}
```

## Testing

### Unit Tests

Unit tests use mock interfaces and do not require an SFTP server:

```bash
go test -v ./...
```

### Integration Tests

Integration tests require a running SFTP server. The project includes Docker Compose configuration for easy setup:

```bash
# Start the SFTP server
docker-compose up -d

# Run integration tests
go test -v -tags=integration ./...

# Stop the server
docker-compose down
```

The Docker setup uses `atmoz/sftp` with the following credentials:
- Host: `localhost:2222`
- Username: `testuser`
- Password: `testpass`

### Benchmarks

Run benchmarks to measure performance:

```bash
go test -bench=. -benchmem ./...
```

### Coverage

Check test coverage:

```bash
go test -cover ./...
```

## Security Note

The current implementation uses `ssh.InsecureIgnoreHostKey()` which skips host key verification. For production use, you should implement proper host key verification to prevent man-in-the-middle attacks.

Example of implementing host key verification:

```go
// For production, implement proper host key callback
hostKeyCallback, err := knownhosts.New("/home/user/.ssh/known_hosts")
if err != nil {
    log.Fatal(err)
}
// Use hostKeyCallback in your SSH configuration
```

## API Reference

### FileSystem Methods

| Method | Description |
|--------|-------------|
| `New(config *Config)` | Create a new SFTP filesystem with configuration |
| `Dial(host, user, password string)` | Quick connect with password auth |
| `DialWithKey(host, user string, privateKey []byte)` | Quick connect with key auth |
| `Close()` | Close the SFTP connection |
| `OpenFile(name string, flag int, perm os.FileMode)` | Open or create a file |
| `Mkdir(name string, perm os.FileMode)` | Create a directory |
| `Remove(name string)` | Remove a file or empty directory |
| `Rename(oldpath, newpath string)` | Rename a file |
| `Stat(name string)` | Get file information |
| `Chmod(name string, mode os.FileMode)` | Change file mode |
| `Chtimes(name string, atime, mtime time.Time)` | Change file times |
| `Chown(name string, uid, gid int)` | Change file ownership |

### File Methods

| Method | Description |
|--------|-------------|
| `Name()` | Return the file name |
| `Read(b []byte)` | Read bytes from file |
| `ReadAt(b []byte, off int64)` | Read at specific offset |
| `Write(b []byte)` | Write bytes to file |
| `WriteAt(b []byte, off int64)` | Write at specific offset |
| `WriteString(s string)` | Write string to file |
| `Seek(offset int64, whence int)` | Seek within file |
| `Close()` | Close the file |
| `Stat()` | Get file information |
| `Sync()` | Sync file (no-op for SFTP) |
| `Truncate(size int64)` | Truncate file to size |
| `Readdir(n int)` | Read directory entries |
| `Readdirnames(n int)` | Read directory entry names |

## absfs

Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract filesystem interface and features like filesystem composition.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/sftpfs/blob/master/LICENSE)
