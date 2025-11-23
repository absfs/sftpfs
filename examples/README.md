# SFTP Filesystem Examples

This directory contains comprehensive examples demonstrating various features and use cases of the `sftpfs` library.

## Prerequisites

All examples require:
- A running SFTP server
- Go 1.21 or later

Set these environment variables before running examples:
```bash
export SFTP_HOST="example.com:22"
export SFTP_USER="username"
export SFTP_PASS="password"
```

For key-based authentication (advanced example):
```bash
export SFTP_KEY_PATH="/home/user/.ssh/id_rsa"
```

## Examples

### 1. Basic Operations (`basic/`)

Demonstrates fundamental filesystem operations:
- Connecting to an SFTP server
- Creating directories
- Writing and reading files
- Getting file information
- Cleaning up resources

**Run:**
```bash
cd examples/basic
go run main.go
```

**What it shows:**
- Simple password authentication
- File creation and reading
- Directory management with `MkdirAll` and `RemoveAll`
- File metadata access with `Stat`

---

### 2. Advanced Configuration (`advanced/`)

Shows production-ready configuration including:
- SSH key authentication with host verification
- Connection retry with exponential backoff
- Multiple authentication methods
- Comprehensive filesystem operations
- Directory tree creation and management

**Run:**
```bash
cd examples/advanced

# With password
SFTP_PASS="password" go run main.go

# With SSH key
SFTP_KEY_PATH="$HOME/.ssh/id_rsa" go run main.go
```

**What it shows:**
- Loading SSH keys from files
- Using `known_hosts` for host key verification
- Retry configuration (5 attempts with 2s initial delay)
- Creating complex directory structures
- Batch file creation
- Permission management with `Chmod`
- Directory listing with `Readdir`

---

### 3. Symbolic Links (`symlinks/`)

Comprehensive demonstration of symlink operations:
- Creating symbolic links
- Reading link targets
- Distinguishing between links and targets
- Handling broken links
- Chained symlinks

**Run:**
```bash
cd examples/symlinks
go run main.go
```

**What it shows:**
- `Symlink` - Create symbolic links
- `Readlink` - Read link targets
- `Lstat` - Get link info without following
- `Stat` - Get target info by following links
- Difference between broken and valid links
- Reading content through symlinks
- Multi-level link chains

---

### 4. Backup Utility (`backup/`)

A practical, production-ready backup tool demonstrating:
- Recursive directory backup from local to remote
- Progress tracking and statistics
- Error handling and recovery
- Timestamped backup directories
- Efficient file transfer with buffering

**Run:**
```bash
cd examples/backup

# Backup current directory to remote
BACKUP_LOCAL_DIR="." BACKUP_REMOTE_DIR="/backups" go run main.go

# Backup specific directory
BACKUP_LOCAL_DIR="/home/user/data" go run main.go
```

**What it shows:**
- Recursive directory walking with `filepath.Walk`
- Preserving directory structure remotely
- File permission preservation
- Large file transfer with buffering
- Real-time progress reporting
- Comprehensive error handling
- Backup statistics (files, bytes, duration)
- Tree-style directory listing

**Features:**
- Skips hidden files (starting with `.`)
- Creates timestamped backup directories
- Shows transfer progress with file sizes
- Displays final statistics
- Lists backup contents in tree format

---

## Common Patterns

### Connection Management

All examples demonstrate proper resource cleanup:
```go
fs, err := sftpfs.Dial(host, user, pass)
if err != nil {
    log.Fatal(err)
}
defer fs.Close() // Always close the connection
```

### Error Handling

Examples show comprehensive error handling:
```go
if err := fs.MkdirAll(path, 0755); err != nil {
    log.Fatalf("Failed to create directory: %v", err)
}
```

### File Operations

Common patterns for file I/O:
```go
// Writing
f, _ := fs.Create("/path/file.txt")
f.Write([]byte("content"))
f.Close()

// Reading
f, _ := fs.Open("/path/file.txt")
data := make([]byte, 1024)
n, _ := f.Read(data)
f.Close()
```

## Tips for Production Use

1. **Always use host key verification** - Don't use `InsecureIgnoreHostKey()` in production
2. **Configure retry logic** - Set appropriate `MaxRetries` and `RetryDelay` for your network
3. **Handle errors gracefully** - Check all error returns and log appropriately
4. **Use buffered I/O** - For large files, use reasonable buffer sizes (32KB-64KB)
5. **Clean up resources** - Always `defer fs.Close()` after connection
6. **Set appropriate permissions** - Use sensible file modes (0644 for files, 0755 for dirs)

## Building Custom Tools

These examples can serve as templates for:
- Backup and restore utilities
- Remote file synchronization tools
- Log collection systems
- Configuration deployment tools
- Remote file monitoring
- Automated file processing pipelines

## Further Reading

- [Main README](../README.md) - Library documentation
- [GoDoc Examples](https://pkg.go.dev/github.com/absfs/sftpfs) - API documentation with examples
- [absfs Documentation](https://github.com/absfs/absfs) - Abstract filesystem interface

## Troubleshooting

**Connection refused:**
```bash
# Check if SFTP server is running
ssh user@host -p 22

# Verify SFTP subsystem is enabled
# In /etc/ssh/sshd_config:
# Subsystem sftp /usr/lib/openssh/sftp-server
```

**Permission denied:**
- Verify user has write permissions to target directories
- Check SSH key permissions (`chmod 600 ~/.ssh/id_rsa`)
- Ensure user has correct ownership of files

**Host key verification failed:**
- Add server to `~/.ssh/known_hosts`: `ssh-keyscan -H hostname >> ~/.ssh/known_hosts`
- Or use the fixed host key method shown in advanced example

## Contributing

Found a bug or want to add an example? Please open an issue or pull request!
