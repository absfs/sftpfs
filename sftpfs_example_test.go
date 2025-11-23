package sftpfs_test

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/absfs/sftpfs"
	"golang.org/x/crypto/ssh"
)

// Example demonstrates basic usage with password authentication
func Example() {
	// Connect using password authentication
	fs, err := sftpfs.Dial("example.com:22", "username", "password")
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	// Create a file
	f, err := fs.Create("/remote/path/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Write to the file
	_, err = f.Write([]byte("Hello, SFTP!"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File created successfully")
}

// ExampleDial shows how to connect with password authentication
func ExampleDial() {
	fs, err := sftpfs.Dial("example.com:22", "username", "password")
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	fmt.Println("Connected successfully")
}

// ExampleDialWithKey shows how to connect with SSH key authentication
func ExampleDialWithKey() {
	// Read your private key
	key, err := os.ReadFile("/home/user/.ssh/id_rsa")
	if err != nil {
		log.Fatal(err)
	}

	fs, err := sftpfs.DialWithKey("example.com:22", "username", key)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	fmt.Println("Connected with key authentication")
}

// ExampleNew shows advanced configuration with all options
func ExampleNew() {
	config := &sftpfs.Config{
		Host:       "example.com:22",
		User:       "username",
		Password:   "password",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
		// For production: provide proper host key callback
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	fs, err := sftpfs.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	fmt.Println("Connected with custom configuration")
}

// ExampleNew_withHostKeyVerification shows production-ready host key verification
func ExampleNew_withHostKeyVerification() {
	// Parse a known host key
	hostKey := "ssh-rsa AAAAB3NzaC1yc2E..."
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(hostKey))
	if err != nil {
		log.Fatal(err)
	}

	config := &sftpfs.Config{
		Host:            "example.com:22",
		User:            "username",
		Password:        "password",
		HostKeyCallback: ssh.FixedHostKey(pubKey),
	}

	fs, err := sftpfs.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	fmt.Println("Connected with host key verification")
}

// ExampleNew_withRetry shows connection retry configuration
func ExampleNew_withRetry() {
	config := &sftpfs.Config{
		Host:       "example.com:22",
		User:       "username",
		Password:   "password",
		MaxRetries: 5,                // Retry up to 5 times
		RetryDelay: 2 * time.Second,  // Start with 2 second delay
		Timeout:    30 * time.Second, // 30 second timeout per attempt
	}

	fs, err := sftpfs.New(config)
	if err != nil {
		log.Fatal(err)
	}
	defer fs.Close()

	fmt.Println("Connected with retry logic")
}

// ExampleFileSystem_Create shows how to create a new file
func ExampleFileSystem_Create() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Create or truncate a file
	f, err := fs.Create("/remote/newfile.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Write data
	f.Write([]byte("Hello, World!"))

	fmt.Println("File created")
}

// ExampleFileSystem_Open shows how to open a file for reading
func ExampleFileSystem_Open() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Open file for reading
	f, err := fs.Open("/remote/file.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Read data
	data := make([]byte, 1024)
	n, err := f.Read(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Read %d bytes\n", n)
}

// ExampleFileSystem_OpenFile shows advanced file opening with flags
func ExampleFileSystem_OpenFile() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Open file with specific flags and permissions
	f, err := fs.OpenFile("/remote/file.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Append data
	f.Write([]byte("Appended text\n"))

	fmt.Println("Data appended")
}

// ExampleFileSystem_Mkdir shows how to create a directory
func ExampleFileSystem_Mkdir() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Create a directory
	err := fs.Mkdir("/remote/newdir", 0755)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Directory created")
}

// ExampleFileSystem_MkdirAll shows how to create nested directories
func ExampleFileSystem_MkdirAll() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Create nested directories in one call
	err := fs.MkdirAll("/remote/path/to/nested/dir", 0755)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Nested directories created")
}

// ExampleFileSystem_Remove shows how to remove a file
func ExampleFileSystem_Remove() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Remove a file
	err := fs.Remove("/remote/file.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File removed")
}

// ExampleFileSystem_RemoveAll shows how to recursively remove directories
func ExampleFileSystem_RemoveAll() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Remove a directory and all its contents
	err := fs.RemoveAll("/remote/directory")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Directory removed")
}

// ExampleFileSystem_Rename shows how to rename or move files
func ExampleFileSystem_Rename() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Rename a file
	err := fs.Rename("/remote/oldname.txt", "/remote/newname.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File renamed")
}

// ExampleFileSystem_Stat shows how to get file information
func ExampleFileSystem_Stat() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Get file info
	info, err := fs.Stat("/remote/file.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Name: %s\n", info.Name())
	fmt.Printf("Size: %d bytes\n", info.Size())
	fmt.Printf("Mode: %s\n", info.Mode())
	fmt.Printf("Modified: %s\n", info.ModTime())
}

// ExampleFileSystem_Chmod shows how to change file permissions
func ExampleFileSystem_Chmod() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Change file permissions
	err := fs.Chmod("/remote/file.txt", 0644)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Permissions changed")
}

// ExampleFileSystem_Chtimes shows how to change file timestamps
func ExampleFileSystem_Chtimes() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Set access and modification times
	now := time.Now()
	err := fs.Chtimes("/remote/file.txt", now, now)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Timestamps updated")
}

// ExampleFileSystem_Chown shows how to change file ownership
func ExampleFileSystem_Chown() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Change file owner and group
	err := fs.Chown("/remote/file.txt", 1000, 1000)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Ownership changed")
}

// ExampleFileSystem_Truncate shows how to truncate a file
func ExampleFileSystem_Truncate() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Truncate file to specific size
	err := fs.Truncate("/remote/file.txt", 1024)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File truncated")
}

// ExampleFileSystem_Chdir shows working directory management
func ExampleFileSystem_Chdir() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Get current directory
	wd, _ := fs.Getwd()
	fmt.Printf("Current directory: %s\n", wd)

	// Change directory
	err := fs.Chdir("/remote/path")
	if err != nil {
		log.Fatal(err)
	}

	// Verify change
	wd, _ = fs.Getwd()
	fmt.Printf("New directory: %s\n", wd)
}

// ExampleFileSystem_Symlink shows symbolic link operations
func ExampleFileSystem_Symlink() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Create a symbolic link
	err := fs.Symlink("/remote/target.txt", "/remote/link.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Symlink created")
}

// ExampleFileSystem_Readlink shows how to read symbolic links
func ExampleFileSystem_Readlink() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Read link target
	target, err := fs.Readlink("/remote/link.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Link points to: %s\n", target)
}

// ExampleFileSystem_Lstat shows how to stat a symlink without following it
func ExampleFileSystem_Lstat() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Get link info without following
	info, err := fs.Lstat("/remote/link.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Link name: %s\n", info.Name())
	fmt.Printf("Link mode: %s\n", info.Mode())
}

// ExampleFileSystem_workingWithDirectories shows comprehensive directory operations
func ExampleFileSystem_workingWithDirectories() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Create directory structure
	fs.MkdirAll("/remote/projects/myapp/src", 0755)
	fs.MkdirAll("/remote/projects/myapp/tests", 0755)

	// Create files
	f1, _ := fs.Create("/remote/projects/myapp/src/main.go")
	f1.Write([]byte("package main\n"))
	f1.Close()

	f2, _ := fs.Create("/remote/projects/myapp/tests/main_test.go")
	f2.Write([]byte("package main\n"))
	f2.Close()

	// List directory contents
	dir, _ := fs.OpenFile("/remote/projects/myapp", os.O_RDONLY, 0)
	entries, _ := dir.Readdir(-1)
	dir.Close()

	for _, entry := range entries {
		fmt.Printf("%s (%d bytes)\n", entry.Name(), entry.Size())
	}

	// Clean up
	fs.RemoveAll("/remote/projects")
}

// ExampleFileSystem_copyFile shows how to copy a file
func ExampleFileSystem_copyFile() {
	fs, _ := sftpfs.Dial("example.com:22", "username", "password")
	defer fs.Close()

	// Open source file
	src, err := fs.Open("/remote/source.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer src.Close()

	// Create destination file
	dst, err := fs.Create("/remote/destination.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer dst.Close()

	// Copy data
	buffer := make([]byte, 32*1024) // 32KB buffer
	for {
		n, err := src.Read(buffer)
		if n > 0 {
			dst.Write(buffer[:n])
		}
		if err != nil {
			break
		}
	}

	fmt.Println("File copied")
}
