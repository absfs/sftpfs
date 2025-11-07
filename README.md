# SftpFs - SFTP FileSystem

[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/absfs/sftpfs/blob/master/LICENSE)

The `sftpfs` package implements an `absfs.Filer` for SFTP (SSH File Transfer Protocol). It provides secure file operations over SSH using the `github.com/pkg/sftp` library.

## Features

- **Secure file operations**: All operations encrypted over SSH
- **Multiple authentication methods**: Password and SSH key authentication
- **Standard interface**: Implements `absfs.Filer` for seamless integration

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

## Security Note

The current implementation uses `ssh.InsecureIgnoreHostKey()` which skips host key verification. For production use, you should implement proper host key verification to prevent man-in-the-middle attacks.

## absfs

Check out the [`absfs`](https://github.com/absfs/absfs) repo for more information about the abstract filesystem interface and features like filesystem composition.

## LICENSE

This project is governed by the MIT License. See [LICENSE](https://github.com/absfs/sftpfs/blob/master/LICENSE)
