# SftpFs - SFTP FileSystem

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/absfs/sftpfs/blob/master/LICENSE)

The `sftpfs` package implements an `absfs.Filer` for SFTP (SSH File Transfer Protocol). It provides secure file operations over SSH using the `github.com/pkg/sftp` library.

## Features

- **Secure file operations**: All operations encrypted over SSH
- **Multiple authentication methods**: Password and SSH key authentication
- **Full filesystem interface**: Implements `absfs.FileSystem` with helper methods (MkdirAll, RemoveAll, Open, Create, etc.)
- **Symbolic link support**: Complete implementation of `absfs.SymLinker` interface
- **Connection resilience**: Automatic retry with exponential backoff for transient failures
- **Enhanced error context**: Detailed error messages with operation and path information
- **Production ready**: Comprehensive test suite with unit and integration tests

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

### Production Configuration with Host Key Verification

For production use, you should implement proper host key verification to prevent man-in-the-middle attacks:

```go
package main

import (
    "log"
    "net"
    "os"

    "github.com/absfs/sftpfs"
    "golang.org/x/crypto/ssh"
    "golang.org/x/crypto/ssh/knownhosts"
)

func main() {
    // Load known hosts file
    hostKeyCallback, err := knownhosts.New(os.ExpandEnv("$HOME/.ssh/known_hosts"))
    if err != nil {
        log.Fatal(err)
    }

    // Read private key
    key, err := os.ReadFile(os.ExpandEnv("$HOME/.ssh/id_rsa"))
    if err != nil {
        log.Fatal(err)
    }

    config := &sftpfs.Config{
        Host:            "example.com:22",
        User:            "username",
        Key:             key,
        HostKeyCallback: hostKeyCallback,
    }

    fs, err := sftpfs.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Use filesystem securely
}
```

Alternatively, for a fixed host key:

```go
// Fixed host key verification
fixedHostKey := "..." // Your server's public key
pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(fixedHostKey))
if err != nil {
    log.Fatal(err)
}

config := &sftpfs.Config{
    Host: "example.com:22",
    User: "username",
    Password: "password",
    HostKeyCallback: ssh.FixedHostKey(pubKey),
}
```

### Connection Retry Configuration

Configure automatic retry with exponential backoff for handling transient network issues:

```go
package main

import (
    "log"
    "time"

    "github.com/absfs/sftpfs"
)

func main() {
    config := &sftpfs.Config{
        Host:       "example.com:22",
        User:       "username",
        Password:   "password",
        MaxRetries: 5,                // Retry up to 5 times (default: 3)
        RetryDelay: 2 * time.Second,  // Initial delay of 2 seconds (default: 1s)
        Timeout:    30 * time.Second, // Connection timeout (default: 30s)
    }

    fs, err := sftpfs.New(config)
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Connection will automatically retry with exponential backoff:
    // Attempt 1: immediate
    // Attempt 2: after 2s
    // Attempt 3: after 4s
    // Attempt 4: after 8s
    // Attempt 5: after 16s
}
```

### Working with Symbolic Links

The filesystem supports full symbolic link operations:

```go
package main

import (
    "log"

    "github.com/absfs/sftpfs"
)

func main() {
    fs, err := sftpfs.Dial("example.com:22", "username", "password")
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Create a symbolic link
    err = fs.Symlink("/path/to/target", "/path/to/link")
    if err != nil {
        log.Fatal(err)
    }

    // Read the link target
    target, err := fs.Readlink("/path/to/link")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Link points to: %s", target)

    // Get link info without following
    info, err := fs.Lstat("/path/to/link")
    if err != nil {
        log.Fatal(err)
    }
    log.Printf("Link info: %v", info)
}
```

### Filesystem Helper Methods

Use convenient helper methods for common operations:

```go
package main

import (
    "log"

    "github.com/absfs/sftpfs"
)

func main() {
    fs, err := sftpfs.Dial("example.com:22", "username", "password")
    if err != nil {
        log.Fatal(err)
    }
    defer fs.Close()

    // Create nested directories in one call
    err = fs.MkdirAll("/path/to/nested/dir", 0755)
    if err != nil {
        log.Fatal(err)
    }

    // Remove a directory and all its contents
    err = fs.RemoveAll("/path/to/directory")
    if err != nil {
        log.Fatal(err)
    }

    // Open file for reading (convenience method)
    f, err := fs.Open("/path/to/file.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer f.Close()

    // Create or truncate a file (convenience method)
    f2, err := fs.Create("/path/to/new_file.txt")
    if err != nil {
        log.Fatal(err)
    }
    f2.Write([]byte("Hello, World!"))
    f2.Close()

    // Working directory operations
    wd, _ := fs.Getwd()
    log.Printf("Current directory: %s", wd)

    fs.Chdir("/home/user")
    wd, _ = fs.Getwd()
    log.Printf("New directory: %s", wd)
}
```

## Testing

Run unit tests:
```bash
go test -v
```

Run integration tests (requires Docker):
```bash
make test-integration
```

Run all tests:
```bash
make test
```

Run benchmarks:
```bash
go test -bench=. -benchmem
```

## Security Note

By default, if no `HostKeyCallback` is provided, the library uses `ssh.InsecureIgnoreHostKey()` which skips host key verification. This is **NOT secure** and vulnerable to man-in-the-middle attacks. Always provide a proper `HostKeyCallback` for production deployments.

## absfs

Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract filesystem interface and features like filesystem composition.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/sftpfs/blob/master/LICENSE)
