// Advanced example showing retry logic, key authentication, and host verification
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/absfs/sftpfs"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

func main() {
	fmt.Println("SFTP Advanced Configuration Example")
	fmt.Println("====================================")
	fmt.Println()

	// Get connection details from environment
	host := os.Getenv("SFTP_HOST")
	user := os.Getenv("SFTP_USER")
	keyPath := os.Getenv("SFTP_KEY_PATH")

	if host == "" || user == "" {
		log.Fatal("Please set SFTP_HOST and SFTP_USER environment variables")
	}

	var config *sftpfs.Config

	// Choose authentication method
	if keyPath != "" {
		// Use SSH key authentication
		fmt.Println("Using SSH key authentication")
		config = configureWithKey(host, user, keyPath)
	} else {
		// Fall back to password
		fmt.Println("Using password authentication")
		pass := os.Getenv("SFTP_PASS")
		if pass == "" {
			log.Fatal("Please set SFTP_PASS or SFTP_KEY_PATH")
		}
		config = configureWithPassword(host, user, pass)
	}

	// Connect with retry logic
	fmt.Println("\nConnecting to SFTP server...")
	start := time.Now()
	fs, err := sftpfs.New(config)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer fs.Close()

	duration := time.Since(start)
	fmt.Printf("✓ Connected successfully in %v\n", duration)

	// Perform some operations
	demonstrateOperations(fs)

	fmt.Println("\n✅ All operations completed successfully!")
}

// configureWithKey sets up SSH key authentication with host verification
func configureWithKey(host, user, keyPath string) *sftpfs.Config {
	// Read the private key
	key, err := os.ReadFile(keyPath)
	if err != nil {
		log.Fatalf("Failed to read key file: %v", err)
	}

	// Load known hosts for host key verification
	knownHostsPath := os.ExpandEnv("$HOME/.ssh/known_hosts")
	hostKeyCallback, err := knownhosts.New(knownHostsPath)
	if err != nil {
		log.Printf("Warning: Could not load known_hosts: %v", err)
		log.Println("Using insecure host key verification")
		hostKeyCallback = ssh.InsecureIgnoreHostKey()
	}

	return &sftpfs.Config{
		Host:            host,
		User:            user,
		Key:             key,
		HostKeyCallback: hostKeyCallback,
		Timeout:         30 * time.Second,
		MaxRetries:      5,               // Retry up to 5 times
		RetryDelay:      2 * time.Second, // Start with 2 second delay
	}
}

// configureWithPassword sets up password authentication
func configureWithPassword(host, user, password string) *sftpfs.Config {
	return &sftpfs.Config{
		Host:       host,
		User:       user,
		Password:   password,
		Timeout:    30 * time.Second,
		MaxRetries: 3,               // Retry up to 3 times
		RetryDelay: 1 * time.Second, // Start with 1 second delay
		// In production, use proper host key verification
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
}

// demonstrateOperations shows various filesystem operations
func demonstrateOperations(fs *sftpfs.FileSystem) {
	fmt.Println("\nPerforming filesystem operations...")

	// Create a test directory structure
	baseDir := "/tmp/sftpfs-advanced-example"
	dirs := []string{
		baseDir + "/data",
		baseDir + "/logs",
		baseDir + "/config",
	}

	for _, dir := range dirs {
		if err := fs.MkdirAll(dir, 0755); err != nil {
			log.Printf("Failed to create %s: %v", dir, err)
			continue
		}
		fmt.Printf("✓ Created: %s\n", dir)
	}

	// Create some files with different content
	files := map[string]string{
		baseDir + "/data/users.json":     `{"users": []}`,
		baseDir + "/logs/app.log":        "Application started\n",
		baseDir + "/config/settings.ini": "[app]\nversion=1.0\n",
	}

	for path, content := range files {
		f, err := fs.Create(path)
		if err != nil {
			log.Printf("Failed to create %s: %v", path, err)
			continue
		}

		if _, err := f.Write([]byte(content)); err != nil {
			log.Printf("Failed to write %s: %v", path, err)
		}
		f.Close()
		fmt.Printf("✓ Created: %s (%d bytes)\n", path, len(content))
	}

	// Change permissions
	configFile := baseDir + "/config/settings.ini"
	if err := fs.Chmod(configFile, 0600); err != nil {
		log.Printf("Failed to chmod: %v", err)
	} else {
		fmt.Printf("✓ Changed permissions: %s -> 0600\n", configFile)
	}

	// List directory contents
	fmt.Printf("\n✓ Contents of %s:\n", baseDir)
	dir, err := fs.OpenFile(baseDir, os.O_RDONLY, 0)
	if err == nil {
		entries, _ := dir.Readdir(-1)
		dir.Close()

		for _, entry := range entries {
			fileType := "file"
			if entry.IsDir() {
				fileType = "dir "
			}
			fmt.Printf("  [%s] %s (%d bytes, %s)\n",
				fileType, entry.Name(), entry.Size(), entry.Mode())
		}
	}

	// Clean up
	fmt.Printf("\n✓ Cleaning up...\n")
	if err := fs.RemoveAll(baseDir); err != nil {
		log.Printf("Failed to clean up: %v", err)
	} else {
		fmt.Printf("✓ Removed: %s\n", baseDir)
	}
}
