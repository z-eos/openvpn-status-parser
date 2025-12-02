package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ServerConfig represents OpenVPN server configuration metadata
type ServerConfig struct {
	// ID is the basename of the status file (without extension)
	ID string `json:"id"`
	
	// Local is the local IP address the server listens on
	Local string `json:"local,omitempty"`
	
	// Port is the port number (default 1194 if not specified)
	Port string `json:"port,omitempty"`
	
	// Proto is the protocol: udp, tcp, udp6, tcp6
	Proto string `json:"proto,omitempty"`
	
	// Dev is the device type: tun or tap
	Dev string `json:"dev,omitempty"`
	
	// StatusFile is the path to the status file
	StatusFile string `json:"-"`
	
	// StatusVersion is the status file format version (1, 2, or 3)
	StatusVersion int `json:"-"`
}

// ParseConfig reads an OpenVPN server configuration file and extracts
// relevant metadata and status file information.
//
// It looks for these directives:
// - local <address>           # Local IP to bind to
// - port <port>               # Port number (default 1194)
// - proto <protocol>          # udp, tcp, udp6, tcp6
// - dev <device>              # tun or tap
// - status <file> [seconds]   # Status file path (we use only the path)
// - status-version <n>        # Status file version: 1, 2, or 3
func ParseConfig(configPath string) (*ServerConfig, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	config := &ServerConfig{
		Port:          "1194", // Default port
		StatusVersion: 3,      // Default to v3 if not specified
	}

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}

		// Split line into tokens (space or tab separated)
		tokens := strings.Fields(line)
		if len(tokens) == 0 {
			continue
		}

		directive := tokens[0]

		switch directive {
		case "local":
			if len(tokens) >= 2 {
				config.Local = tokens[1]
			}

		case "port":
			if len(tokens) >= 2 {
				config.Port = tokens[1]
			}

		case "proto":
			if len(tokens) >= 2 {
				config.Proto = tokens[1]
			}

		case "dev":
			if len(tokens) >= 2 {
				config.Dev = tokens[1]
			}

		case "status":
			if len(tokens) >= 2 {
				// Extract only the first argument (file path)
				// Ignore second argument (refresh interval) if present
				config.StatusFile = tokens[1]
				
				// Generate server ID from status file basename
				config.ID = getServerID(config.StatusFile)
			}

		case "status-version":
			if len(tokens) >= 2 {
				if ver, err := strconv.Atoi(tokens[1]); err == nil {
					if ver >= 1 && ver <= 3 {
						config.StatusVersion = ver
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Validate that we found a status file
	if config.StatusFile == "" {
		return nil, fmt.Errorf("no 'status' directive found in config file")
	}

	return config, nil
}

// getServerID extracts a server identifier from the status file path.
// It returns the basename without extension.
// Example: /var/log/openvpn/status.log -> "status"
// Example: /var/log/openvpn/server1-status.log -> "server1-status"
func getServerID(statusPath string) string {
	// Get basename
	basename := filepath.Base(statusPath)
	
	// Remove extension
	ext := filepath.Ext(basename)
	if ext != "" {
		basename = basename[:len(basename)-len(ext)]
	}
	
	return basename
}
