# SftpFs - SFTP FileSystem

[![Go Reference](https://pkg.go.dev/badge/github.com/absfs/sftpfs.svg)](https://pkg.go.dev/github.com/absfs/sftpfs)
[![Go Report Card](https://goreportcard.com/badge/github.com/absfs/sftpfs)](https://goreportcard.com/report/github.com/absfs/sftpfs)
[![CI](https://github.com/absfs/sftpfs/actions/workflows/ci.yml/badge.svg)](https://github.com/absfs/sftpfs/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

The `sftpfs` package provides bidirectional SFTP support for the absfs ecosystem:

- **Client Mode**: Access remote SFTP servers as an `absfs.FileSystem`
- **Server Mode**: Serve any `absfs.FileSystem` over SFTP protocol

## Features

- **Bidirectional**: Both client and server implementations
- **Secure file operations**: All operations encrypted over SSH
- **Multiple authentication methods**: Password and SSH key authentication
- **Standard interface**: Client implements `absfs.Filer` for seamless integration
- **Full file operations**: Read, write, seek, truncate, and more
- **Directory operations**: Create, remove, and list directories
- **Server mode**: Expose any absfs filesystem via SFTP

## Install

```bash
go get github.com/absfs/sftpfs
```

## Client Usage

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

## Server Usage

The server mode allows you to expose any `absfs.FileSystem` over SFTP protocol. This is useful for creating custom file servers, testing, or bridging different storage backends.

### Basic Server

```go
package main

import (
    "crypto/rand"
    "crypto/rsa"
    "log"
    "net"

    "github.com/absfs/memfs"
    "github.com/absfs/sftpfs"
    "golang.org/x/crypto/ssh"
)

func main() {
    // Create a filesystem to serve (could be any absfs.FileSystem)
    fs, err := memfs.NewFS()
    if err != nil {
        log.Fatal(err)
    }

    // Generate a host key (in production, load from file)
    privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
    if err != nil {
        log.Fatal(err)
    }
    signer, err := ssh.NewSignerFromKey(privateKey)
    if err != nil {
        log.Fatal(err)
    }

    // Create SFTP server with password authentication
    server := sftpfs.NewServer(fs, &sftpfs.ServerConfig{
        HostKeys: []ssh.Signer{signer},
        PasswordCallback: sftpfs.SimplePasswordAuth("admin", "secret"),
    })

    // Listen and serve
    listener, err := net.Listen("tcp", ":2222")
    if err != nil {
        log.Fatal(err)
    }
    log.Println("SFTP server listening on :2222")
    log.Fatal(server.Serve(listener))
}
```

### Multi-User Authentication

```go
users := map[string]string{
    "alice": "password1",
    "bob":   "password2",
    "carol": "password3",
}

server := sftpfs.NewServer(fs, &sftpfs.ServerConfig{
    HostKeys:         []ssh.Signer{signer},
    PasswordCallback: sftpfs.MultiUserPasswordAuth(users),
})
```

### Public Key Authentication

```go
server := sftpfs.NewServer(fs, &sftpfs.ServerConfig{
    HostKeys: []ssh.Signer{hostKey},
    PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
        // Verify the public key against your authorized keys
        authorizedKey, _, _, _, err := ssh.ParseAuthorizedKey(authorizedKeysData)
        if err != nil {
            return nil, err
        }

        if ssh.KeysEqual(key, authorizedKey) {
            return nil, nil // Authentication successful
        }
        return nil, fmt.Errorf("unknown public key")
    },
})
```

### Serving Different Filesystems

```go
// Serve local files
osfs, _ := osfs.NewFS()
server := sftpfs.NewServer(osfs, config)

// Serve in-memory files
memfs, _ := memfs.NewFS()
server := sftpfs.NewServer(memfs, config)

// Serve composed filesystems
union := unionfs.New(baseFS, overlayFS)
server := sftpfs.NewServer(union, config)
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

### Client Types and Methods

#### FileSystem Methods

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

#### File Methods

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

### Server Types and Methods

#### Server

| Method | Description |
|--------|-------------|
| `NewServer(fs absfs.FileSystem, config *ServerConfig)` | Create a new SFTP server |
| `Serve(listener net.Listener)` | Accept connections and serve SFTP |
| `ServeConn(conn net.Conn)` | Handle a single connection |
| `SSHConfig()` | Get the underlying SSH server config |

#### ServerConfig

| Field | Type | Description |
|-------|------|-------------|
| `HostKeys` | `[]ssh.Signer` | SSH host keys (at least one required) |
| `PasswordCallback` | `func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error)` | Password authentication handler |
| `PublicKeyCallback` | `func(ssh.ConnMetadata, ssh.PublicKey) (*ssh.Permissions, error)` | Public key authentication handler |
| `NoClientAuth` | `bool` | Allow connections without authentication (testing only) |
| `MaxAuthTries` | `int` | Maximum authentication attempts (default: 6) |
| `ServerVersion` | `string` | SSH server version string |

#### Helper Functions

| Function | Description |
|----------|-------------|
| `SimplePasswordAuth(user, pass string)` | Create single-user password callback |
| `MultiUserPasswordAuth(users map[string]string)` | Create multi-user password callback |
| `NewServerHandler(fs absfs.FileSystem)` | Create low-level SFTP handlers |

## absfs

Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract filesystem interface and features like filesystem composition.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/sftpfs/blob/master/LICENSE)
