// Basic example showing simple file operations with sftpfs
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/absfs/sftpfs"
)

func main() {
	// Connect to SFTP server
	// Replace with your actual server details
	host := os.Getenv("SFTP_HOST")
	user := os.Getenv("SFTP_USER")
	pass := os.Getenv("SFTP_PASS")

	if host == "" || user == "" || pass == "" {
		log.Fatal("Please set SFTP_HOST, SFTP_USER, and SFTP_PASS environment variables")
	}

	fs, err := sftpfs.Dial(host, user, pass)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer fs.Close()

	fmt.Println("✓ Connected to SFTP server")

	// Create a test directory
	testDir := "/tmp/sftpfs-example"
	if err := fs.MkdirAll(testDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	fmt.Printf("✓ Created directory: %s\n", testDir)

	// Create and write to a file
	testFile := testDir + "/hello.txt"
	f, err := fs.Create(testFile)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}

	message := "Hello from sftpfs!\n"
	if _, err := f.Write([]byte(message)); err != nil {
		log.Fatalf("Failed to write: %v", err)
	}
	f.Close()
	fmt.Printf("✓ Created file: %s\n", testFile)

	// Read the file back
	f, err = fs.Open(testFile)
	if err != nil {
		log.Fatalf("Failed to open file: %v", err)
	}

	data := make([]byte, 100)
	n, err := f.Read(data)
	if err != nil {
		log.Fatalf("Failed to read: %v", err)
	}
	f.Close()

	fmt.Printf("✓ Read %d bytes: %s", n, string(data[:n]))

	// Get file info
	info, err := fs.Stat(testFile)
	if err != nil {
		log.Fatalf("Failed to stat: %v", err)
	}

	fmt.Printf("✓ File info:\n")
	fmt.Printf("  Name: %s\n", info.Name())
	fmt.Printf("  Size: %d bytes\n", info.Size())
	fmt.Printf("  Mode: %s\n", info.Mode())
	fmt.Printf("  Modified: %s\n", info.ModTime())

	// Clean up
	if err := fs.RemoveAll(testDir); err != nil {
		log.Fatalf("Failed to clean up: %v", err)
	}
	fmt.Printf("✓ Cleaned up: %s\n", testDir)

	fmt.Println("\n✅ All operations completed successfully!")
}
