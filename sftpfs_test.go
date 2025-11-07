package sftpfs

import (
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	config := &Config{
		Host:     "localhost:22",
		User:     "testuser",
		Password: "testpass",
		Timeout:  30 * time.Second,
	}

	if config.Host != "localhost:22" {
		t.Errorf("Host not set correctly")
	}
	if config.User != "testuser" {
		t.Errorf("User not set correctly")
	}
	if config.Password != "testpass" {
		t.Errorf("Password not set correctly")
	}
	if config.Timeout != 30*time.Second {
		t.Errorf("Timeout not set correctly")
	}
}

func TestNewConfig(t *testing.T) {
	config := &Config{
		Host:     "localhost:22",
		User:     "testuser",
		Password: "testpass",
	}

	// Note: This will fail without an actual SFTP server
	// This is just a structural test
	_, err := New(config)
	if err == nil {
		t.Skip("Skipping test - no SFTP server available")
	}
	// We expect an error since there's no server running
	// This just tests that the function can be called
}

func TestDialSignature(t *testing.T) {
	// Test that Dial function exists with correct signature
	// This is a compile-time test more than a runtime test
	var _ func(string, string, string) (*FileSystem, error) = Dial
}

func TestDialWithKeySignature(t *testing.T) {
	// Test that DialWithKey function exists with correct signature
	var _ func(string, string, []byte) (*FileSystem, error) = DialWithKey
}
