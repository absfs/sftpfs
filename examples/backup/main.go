// Example showing a practical backup utility using sftpfs
package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/absfs/sftpfs"
)

func main() {
	fmt.Println("SFTP Backup Utility Example")
	fmt.Println("============================")
	fmt.Println()

	// Get configuration from environment
	host := os.Getenv("SFTP_HOST")
	user := os.Getenv("SFTP_USER")
	pass := os.Getenv("SFTP_PASS")
	localDir := os.Getenv("BACKUP_LOCAL_DIR")
	remoteDir := os.Getenv("BACKUP_REMOTE_DIR")

	if host == "" || user == "" || pass == "" {
		log.Fatal("Please set SFTP_HOST, SFTP_USER, and SFTP_PASS")
	}

	if localDir == "" {
		localDir = "."
	}
	if remoteDir == "" {
		remoteDir = "/tmp/backups"
	}

	// Connect to SFTP server
	fmt.Printf("Connecting to %s...\n", host)
	fs, err := sftpfs.Dial(host, user, pass)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer fs.Close()
	fmt.Println("✓ Connected successfully")
	fmt.Println()

	// Create timestamped backup directory
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	backupDir := path.Join(remoteDir, "backup_"+timestamp)

	fmt.Printf("Creating backup directory: %s\n", backupDir)
	if err := fs.MkdirAll(backupDir, 0755); err != nil {
		log.Fatalf("Failed to create backup directory: %v", err)
	}

	// Backup files
	fmt.Printf("\nBacking up files from %s to %s...\n", localDir, backupDir)
	stats := backupDirectory(fs, localDir, backupDir)

	// Print statistics
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Backup Summary")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Files backed up:  %d\n", stats.filesBackedUp)
	fmt.Printf("Bytes transferred: %d (%.2f MB)\n",
		stats.bytesTransferred,
		float64(stats.bytesTransferred)/(1024*1024))
	fmt.Printf("Directories created: %d\n", stats.dirsCreated)
	fmt.Printf("Errors encountered: %d\n", stats.errors)
	fmt.Printf("Duration: %v\n", stats.duration)
	fmt.Println(strings.Repeat("=", 50))

	// List backup contents
	fmt.Printf("\nBackup contents:\n")
	listDirectoryRecursive(fs, backupDir, "  ")

	fmt.Println("\n✅ Backup completed successfully!")
	fmt.Printf("Remote backup location: %s\n", backupDir)
}

// BackupStats tracks backup operation statistics
type BackupStats struct {
	filesBackedUp    int
	bytesTransferred int64
	dirsCreated      int
	errors           int
	duration         time.Duration
}

// backupDirectory recursively backs up a local directory to remote SFTP
func backupDirectory(fs *sftpfs.FileSystem, localPath, remotePath string) BackupStats {
	start := time.Now()
	stats := BackupStats{}

	err := filepath.Walk(localPath, func(localFilePath string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("Error accessing %s: %v", localFilePath, err)
			stats.errors++
			return nil
		}

		// Skip hidden files and directories
		if filepath.Base(localFilePath)[0] == '.' {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(localPath, localFilePath)
		if err != nil {
			log.Printf("Error calculating relative path: %v", err)
			stats.errors++
			return nil
		}

		// Convert to remote path (use forward slashes)
		remoteFilePath := path.Join(remotePath, filepath.ToSlash(relPath))

		if info.IsDir() {
			// Create remote directory
			if err := fs.MkdirAll(remoteFilePath, info.Mode().Perm()); err != nil {
				log.Printf("Error creating directory %s: %v", remoteFilePath, err)
				stats.errors++
				return nil
			}
			stats.dirsCreated++
			fmt.Printf("  📁 %s\n", relPath)
		} else {
			// Copy file
			bytes, err := copyFile(fs, localFilePath, remoteFilePath, info.Mode().Perm())
			if err != nil {
				log.Printf("Error copying %s: %v", localFilePath, err)
				stats.errors++
				return nil
			}
			stats.filesBackedUp++
			stats.bytesTransferred += bytes
			fmt.Printf("  📄 %s (%d bytes)\n", relPath, bytes)
		}

		return nil
	})

	if err != nil {
		log.Printf("Walk error: %v", err)
		stats.errors++
	}

	stats.duration = time.Since(start)
	return stats
}

// copyFile copies a single file from local to remote
func copyFile(fs *sftpfs.FileSystem, localPath, remotePath string, perm os.FileMode) (int64, error) {
	// Open local file
	localFile, err := os.Open(localPath)
	if err != nil {
		return 0, fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Create remote file
	remoteFile, err := fs.OpenFile(remotePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return 0, fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	// Copy data
	buffer := make([]byte, 64*1024) // 64KB buffer
	var totalBytes int64

	for {
		n, err := localFile.Read(buffer)
		if n > 0 {
			written, writeErr := remoteFile.Write(buffer[:n])
			if writeErr != nil {
				return totalBytes, fmt.Errorf("failed to write: %w", writeErr)
			}
			totalBytes += int64(written)
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return totalBytes, fmt.Errorf("failed to read: %w", err)
		}
	}

	return totalBytes, nil
}

// listDirectoryRecursive lists directory contents recursively
func listDirectoryRecursive(fs *sftpfs.FileSystem, dirPath, prefix string) {
	dir, err := fs.OpenFile(dirPath, os.O_RDONLY, 0)
	if err != nil {
		log.Printf("Failed to open directory %s: %v", dirPath, err)
		return
	}
	defer dir.Close()

	entries, err := dir.Readdir(-1)
	if err != nil {
		log.Printf("Failed to read directory %s: %v", dirPath, err)
		return
	}

	for i, entry := range entries {
		isLast := i == len(entries)-1
		connector := "├──"
		if isLast {
			connector = "└──"
		}

		icon := "📄"
		if entry.IsDir() {
			icon = "📁"
		}

		fmt.Printf("%s%s %s %s (%d bytes)\n",
			prefix, connector, icon, entry.Name(), entry.Size())

		if entry.IsDir() {
			childPrefix := prefix
			if isLast {
				childPrefix += "    "
			} else {
				childPrefix += "│   "
			}
			childPath := path.Join(dirPath, entry.Name())
			listDirectoryRecursive(fs, childPath, childPrefix)
		}
	}
}
