package sftpfs

import (
	"net"

	"github.com/absfs/absfs"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

// Server provides an SFTP server backed by any absfs.FileSystem.
// It handles SSH connections and SFTP protocol negotiation.
type Server struct {
	fs       absfs.FileSystem
	config   *ssh.ServerConfig
	handlers sftp.Handlers
}

// ServerConfig holds configuration for the SFTP server.
type ServerConfig struct {
	// HostKeys are the private keys for the SSH server.
	// At least one host key is required.
	HostKeys []ssh.Signer

	// PasswordCallback validates password authentication.
	// If nil, password authentication is disabled.
	PasswordCallback func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error)

	// PublicKeyCallback validates public key authentication.
	// If nil, public key authentication is disabled.
	PublicKeyCallback func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error)

	// NoClientAuth allows any client to connect without authentication.
	// WARNING: Only use this for testing or trusted networks.
	NoClientAuth bool

	// MaxAuthTries specifies the maximum number of authentication attempts.
	// If 0, defaults to 6.
	MaxAuthTries int

	// ServerVersion is the SSH server version string.
	// If empty, defaults to "SSH-2.0-sftpfs".
	ServerVersion string
}

// NewServer creates a new SFTP server for the given filesystem.
//
// Example usage:
//
//	fs, _ := memfs.NewFS()
//
//	// Load or generate host key
//	privateKey, _ := rsa.GenerateKey(rand.Reader, 2048)
//	signer, _ := ssh.NewSignerFromKey(privateKey)
//
//	server := sftpfs.NewServer(fs, &sftpfs.ServerConfig{
//	    HostKeys: []ssh.Signer{signer},
//	    PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
//	        if c.User() == "admin" && string(pass) == "secret" {
//	            return nil, nil
//	        }
//	        return nil, fmt.Errorf("invalid credentials")
//	    },
//	})
//
//	listener, _ := net.Listen("tcp", ":2222")
//	server.Serve(listener)
func NewServer(fs absfs.FileSystem, config *ServerConfig) *Server {
	if config == nil {
		config = &ServerConfig{}
	}

	sshConfig := &ssh.ServerConfig{}

	// Configure authentication
	if config.NoClientAuth {
		sshConfig.NoClientAuth = true
	} else {
		if config.PasswordCallback != nil {
			sshConfig.PasswordCallback = config.PasswordCallback
		}
		if config.PublicKeyCallback != nil {
			sshConfig.PublicKeyCallback = config.PublicKeyCallback
		}
	}

	// Add host keys
	for _, key := range config.HostKeys {
		sshConfig.AddHostKey(key)
	}

	// Set max auth tries
	if config.MaxAuthTries > 0 {
		sshConfig.MaxAuthTries = config.MaxAuthTries
	}

	// Set server version
	if config.ServerVersion != "" {
		sshConfig.ServerVersion = config.ServerVersion
	} else {
		sshConfig.ServerVersion = "SSH-2.0-sftpfs"
	}

	return &Server{
		fs:       fs,
		config:   sshConfig,
		handlers: NewServerHandler(fs),
	}
}

// Serve accepts incoming connections on the listener and serves SFTP.
// This function blocks until the listener is closed.
func (s *Server) Serve(listener net.Listener) error {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go s.handleConnection(conn)
	}
}

// ServeConn handles a single incoming connection.
// This is useful for custom connection handling or testing.
func (s *Server) ServeConn(conn net.Conn) error {
	return s.handleConnection(conn)
}

// handleConnection performs SSH handshake and serves SFTP.
func (s *Server) handleConnection(conn net.Conn) error {
	// Perform SSH handshake
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		conn.Close()
		return err
	}
	defer sshConn.Close()

	// Discard global requests
	go ssh.DiscardRequests(reqs)

	// Handle channels
	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}

		go s.handleChannel(channel, requests)
	}

	return nil
}

// handleChannel handles an SSH channel, looking for SFTP subsystem requests.
func (s *Server) handleChannel(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer channel.Close()

	for req := range requests {
		ok := false
		switch req.Type {
		case "subsystem":
			if string(req.Payload[4:]) == "sftp" {
				ok = true
				if req.WantReply {
					req.Reply(ok, nil)
				}
				s.serveSFTP(channel)
				return
			}
		}
		if req.WantReply {
			req.Reply(ok, nil)
		}
	}
}

// serveSFTP creates and runs an SFTP server on the channel.
func (s *Server) serveSFTP(channel ssh.Channel) {
	server := sftp.NewRequestServer(channel, s.handlers)
	server.Serve()
	server.Close()
}

// SSHConfig returns the underlying SSH server configuration.
// This can be used to add additional configuration options.
func (s *Server) SSHConfig() *ssh.ServerConfig {
	return s.config
}

// SimplePasswordAuth returns a PasswordCallback that validates a single user/password.
// This is a convenience function for simple authentication scenarios.
func SimplePasswordAuth(username, password string) func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		if conn.User() == username && string(pass) == password {
			return nil, nil
		}
		return nil, ErrAuthFailed
	}
}

// MultiUserPasswordAuth returns a PasswordCallback that validates against a user/password map.
func MultiUserPasswordAuth(users map[string]string) func(ssh.ConnMetadata, []byte) (*ssh.Permissions, error) {
	return func(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
		expectedPass, ok := users[conn.User()]
		if ok && expectedPass == string(pass) {
			return nil, nil
		}
		return nil, ErrAuthFailed
	}
}

// ErrAuthFailed is returned when authentication fails.
var ErrAuthFailed = &AuthError{msg: "authentication failed"}

// AuthError represents an authentication failure.
type AuthError struct {
	msg string
}

func (e *AuthError) Error() string {
	return e.msg
}
