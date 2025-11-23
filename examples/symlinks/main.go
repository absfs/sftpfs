// Example demonstrating symbolic link operations
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/absfs/sftpfs"
)

func main() {
	fmt.Println("SFTP Symbolic Links Example")
	fmt.Println("============================")
	fmt.Println()

	// Connect to SFTP server
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

	// Create test directory
	testDir := "/tmp/sftpfs-symlinks-example"
	if err := fs.MkdirAll(testDir, 0755); err != nil {
		log.Fatalf("Failed to create directory: %v", err)
	}
	fmt.Printf("✓ Created test directory: %s\n\n", testDir)

	// Demonstrate various symlink operations
	demonstrateBasicSymlinks(fs, testDir)
	demonstrateSymlinkChains(fs, testDir)
	demonstrateBrokenSymlinks(fs, testDir)

	// Clean up
	if err := fs.RemoveAll(testDir); err != nil {
		log.Printf("Failed to clean up: %v", err)
	} else {
		fmt.Printf("\n✓ Cleaned up test directory\n")
	}

	fmt.Println("\n✅ All symlink operations completed!")
}

// demonstrateBasicSymlinks shows basic symbolic link creation and reading
func demonstrateBasicSymlinks(fs *sftpfs.FileSystem, baseDir string) {
	fmt.Println("1. Basic Symbolic Links")
	fmt.Println("   -------------------")

	// Create a target file
	targetFile := baseDir + "/original.txt"
	f, err := fs.Create(targetFile)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	f.Write([]byte("This is the original file content.\n"))
	f.Close()
	fmt.Printf("   ✓ Created target file: %s\n", targetFile)

	// Create a symbolic link to the file
	linkFile := baseDir + "/link-to-original.txt"
	if err := fs.Symlink(targetFile, linkFile); err != nil {
		log.Fatalf("Failed to create symlink: %v", err)
	}
	fmt.Printf("   ✓ Created symlink: %s -> %s\n", linkFile, targetFile)

	// Read the link target
	target, err := fs.Readlink(linkFile)
	if err != nil {
		log.Fatalf("Failed to read symlink: %v", err)
	}
	fmt.Printf("   ✓ Readlink returned: %s\n", target)

	// Use Lstat to get link info (without following)
	linkInfo, err := fs.Lstat(linkFile)
	if err != nil {
		log.Fatalf("Failed to lstat: %v", err)
	}
	fmt.Printf("   ✓ Lstat (link itself): %s, mode=%s\n",
		linkInfo.Name(), linkInfo.Mode())

	// Use Stat to get target info (following the link)
	targetInfo, err := fs.Stat(linkFile)
	if err != nil {
		log.Fatalf("Failed to stat: %v", err)
	}
	fmt.Printf("   ✓ Stat (target file): size=%d bytes\n", targetInfo.Size())

	// Read content through the symlink
	f, err = fs.Open(linkFile)
	if err != nil {
		log.Fatalf("Failed to open symlink: %v", err)
	}
	data := make([]byte, 100)
	n, _ := f.Read(data)
	f.Close()
	fmt.Printf("   ✓ Read through symlink: %s", string(data[:n]))

	fmt.Println()
}

// demonstrateSymlinkChains shows chained symbolic links
func demonstrateSymlinkChains(fs *sftpfs.FileSystem, baseDir string) {
	fmt.Println("2. Chained Symbolic Links")
	fmt.Println("   ----------------------")

	// Create a directory and file
	dataDir := baseDir + "/data"
	fs.Mkdir(dataDir, 0755)

	versionedFile := dataDir + "/app-v1.0.txt"
	f, _ := fs.Create(versionedFile)
	f.Write([]byte("Version 1.0 content\n"))
	f.Close()
	fmt.Printf("   ✓ Created: %s\n", versionedFile)

	// Create first link: app-latest -> app-v1.0.txt
	latestLink := dataDir + "/app-latest.txt"
	fs.Symlink(versionedFile, latestLink)
	fmt.Printf("   ✓ Created: %s -> %s\n", latestLink, versionedFile)

	// Create second link: app-current -> app-latest.txt
	currentLink := dataDir + "/app-current.txt"
	fs.Symlink(latestLink, currentLink)
	fmt.Printf("   ✓ Created: %s -> %s\n", currentLink, latestLink)

	// Read through the chain
	target1, _ := fs.Readlink(currentLink)
	fmt.Printf("   ✓ %s points to: %s\n", currentLink, target1)

	target2, _ := fs.Readlink(latestLink)
	fmt.Printf("   ✓ %s points to: %s\n", latestLink, target2)

	// Read final content
	f, _ = fs.Open(currentLink)
	data := make([]byte, 100)
	n, _ := f.Read(data)
	f.Close()
	fmt.Printf("   ✓ Final content: %s", string(data[:n]))

	fmt.Println()
}

// demonstrateBrokenSymlinks shows handling of broken symbolic links
func demonstrateBrokenSymlinks(fs *sftpfs.FileSystem, baseDir string) {
	fmt.Println("3. Broken Symbolic Links")
	fmt.Println("   --------------------")

	// Create a link to a non-existent file
	brokenLink := baseDir + "/broken-link.txt"
	nonExistent := baseDir + "/does-not-exist.txt"

	if err := fs.Symlink(nonExistent, brokenLink); err != nil {
		log.Printf("   Failed to create broken link: %v", err)
		return
	}
	fmt.Printf("   ✓ Created broken link: %s -> %s\n", brokenLink, nonExistent)

	// Lstat works on broken links
	info, err := fs.Lstat(brokenLink)
	if err != nil {
		log.Printf("   Lstat failed: %v", err)
	} else {
		fmt.Printf("   ✓ Lstat (broken link): %s, mode=%s\n",
			info.Name(), info.Mode())
	}

	// Readlink works on broken links
	target, err := fs.Readlink(brokenLink)
	if err != nil {
		log.Printf("   Readlink failed: %v", err)
	} else {
		fmt.Printf("   ✓ Readlink: %s\n", target)
	}

	// Stat fails on broken links (tries to follow)
	_, err = fs.Stat(brokenLink)
	if err != nil {
		fmt.Printf("   ✓ Stat correctly fails on broken link: %v\n", err)
	} else {
		fmt.Printf("   ⚠ Unexpected: Stat succeeded on broken link\n")
	}

	// Open fails on broken links
	_, err = fs.Open(brokenLink)
	if err != nil {
		fmt.Printf("   ✓ Open correctly fails on broken link: %v\n", err)
	} else {
		fmt.Printf("   ⚠ Unexpected: Open succeeded on broken link\n")
	}

	// Now create the target to "fix" the link
	f, _ := fs.Create(nonExistent)
	f.Write([]byte("Now the link works!\n"))
	f.Close()
	fmt.Printf("   ✓ Created target file, link is now valid\n")

	// Now operations should work
	f, err = fs.Open(brokenLink)
	if err != nil {
		log.Printf("   Open still failed: %v", err)
	} else {
		data := make([]byte, 100)
		n, _ := f.Read(data)
		f.Close()
		fmt.Printf("   ✓ Content through fixed link: %s", string(data[:n]))
	}

	fmt.Println()
}
