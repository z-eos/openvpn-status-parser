package config

import (
	"os"
	"testing"
)

// TestParseConfigComplete tests parsing a complete config file
func TestParseConfigComplete(t *testing.T) {
	content := `# OpenVPN Server Configuration
local 192.168.1.100
port 1194
proto udp
dev tun
status /var/log/openvpn/status.log 30
status-version 3

# Other directives
ca /etc/openvpn/ca.crt
cert /etc/openvpn/server.crt
key /etc/openvpn/server.key`

	tmpfile := createTempFile(t, "server-*.conf", content)
	defer os.Remove(tmpfile)

	config, err := ParseConfig(tmpfile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.Local != "192.168.1.100" {
		t.Errorf("Expected Local '192.168.1.100', got '%s'", config.Local)
	}
	if config.Port != "1194" {
		t.Errorf("Expected Port '1194', got '%s'", config.Port)
	}
	if config.Proto != "udp" {
		t.Errorf("Expected Proto 'udp', got '%s'", config.Proto)
	}
	if config.Dev != "tun" {
		t.Errorf("Expected Dev 'tun', got '%s'", config.Dev)
	}
	if config.StatusFile != "/var/log/openvpn/status.log" {
		t.Errorf("Expected StatusFile '/var/log/openvpn/status.log', got '%s'", config.StatusFile)
	}
	if config.StatusVersion != 3 {
		t.Errorf("Expected StatusVersion 3, got %d", config.StatusVersion)
	}
	if config.ID != "status" {
		t.Errorf("Expected ID 'status', got '%s'", config.ID)
	}
}

// TestParseConfigMinimal tests parsing with minimal required directives
func TestParseConfigMinimal(t *testing.T) {
	content := `status /var/log/openvpn/status.log`

	tmpfile := createTempFile(t, "server-minimal-*.conf", content)
	defer os.Remove(tmpfile)

	config, err := ParseConfig(tmpfile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.Port != "1194" {
		t.Errorf("Expected default Port '1194', got '%s'", config.Port)
	}
	if config.StatusVersion != 3 {
		t.Errorf("Expected default StatusVersion 3, got %d", config.StatusVersion)
	}
	if config.StatusFile != "/var/log/openvpn/status.log" {
		t.Errorf("Expected StatusFile '/var/log/openvpn/status.log', got '%s'", config.StatusFile)
	}
}

// TestParseConfigNoStatus tests error when no status directive found
func TestParseConfigNoStatus(t *testing.T) {
	content := `local 192.168.1.100
port 1194
proto udp`

	tmpfile := createTempFile(t, "server-nostatus-*.conf", content)
	defer os.Remove(tmpfile)

	_, err := ParseConfig(tmpfile)
	if err == nil {
		t.Error("Expected error when status directive is missing, got none")
	}
}

// TestParseConfigInvalidPath tests error handling for non-existent file
func TestParseConfigInvalidPath(t *testing.T) {
	_, err := ParseConfig("/nonexistent/path/server.conf")
	if err == nil {
		t.Error("Expected error for non-existent file, got none")
	}
}

// TestParseConfigComments tests that comments are ignored
func TestParseConfigComments(t *testing.T) {
	content := `# This is a comment
; This is also a comment
local 192.168.1.100
# port 1195
port 1194
status /var/log/openvpn/status.log`

	tmpfile := createTempFile(t, "server-comments-*.conf", content)
	defer os.Remove(tmpfile)

	config, err := ParseConfig(tmpfile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.Port != "1194" {
		t.Errorf("Expected Port '1194', got '%s'", config.Port)
	}
}

// TestParseConfigStatusWithRefresh tests status directive with refresh interval
func TestParseConfigStatusWithRefresh(t *testing.T) {
	content := `status /var/log/openvpn/status.log 60`

	tmpfile := createTempFile(t, "server-refresh-*.conf", content)
	defer os.Remove(tmpfile)

	config, err := ParseConfig(tmpfile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.StatusFile != "/var/log/openvpn/status.log" {
		t.Errorf("Expected StatusFile '/var/log/openvpn/status.log', got '%s'", config.StatusFile)
	}
}

// TestParseConfigStatusVersion tests all three version values
func TestParseConfigStatusVersion(t *testing.T) {
	tests := []struct {
		version  string
		expected int
	}{
		{"1", 1},
		{"2", 2},
		{"3", 3},
	}

	for _, tt := range tests {
		content := "status /var/log/openvpn/status.log\nstatus-version " + tt.version

		tmpfile := createTempFile(t, "server-version-*.conf", content)
		config, err := ParseConfig(tmpfile)
		os.Remove(tmpfile)

		if err != nil {
			t.Fatalf("ParseConfig failed for version %s: %v", tt.version, err)
		}

		if config.StatusVersion != tt.expected {
			t.Errorf("Expected StatusVersion %d, got %d", tt.expected, config.StatusVersion)
		}
	}
}

// TestParseConfigStatusVersionInvalid tests invalid status-version values
func TestParseConfigStatusVersionInvalid(t *testing.T) {
	tests := []string{"0", "4", "invalid", "-1"}

	for _, version := range tests {
		content := "status /var/log/openvpn/status.log\nstatus-version " + version

		tmpfile := createTempFile(t, "server-bad-version-*.conf", content)
		config, err := ParseConfig(tmpfile)
		os.Remove(tmpfile)

		if err != nil {
			t.Fatalf("ParseConfig failed: %v", err)
		}

		if config.StatusVersion != 3 {
			t.Errorf("Expected default StatusVersion 3 for invalid value '%s', got %d", version, config.StatusVersion)
		}
	}
}

// TestParseConfigProtocols tests different protocol values
func TestParseConfigProtocols(t *testing.T) {
	protocols := []string{"udp", "tcp", "udp6", "tcp6"}

	for _, proto := range protocols {
		content := "status /var/log/openvpn/status.log\nproto " + proto

		tmpfile := createTempFile(t, "server-proto-*.conf", content)
		config, err := ParseConfig(tmpfile)
		os.Remove(tmpfile)

		if err != nil {
			t.Fatalf("ParseConfig failed for proto %s: %v", proto, err)
		}

		if config.Proto != proto {
			t.Errorf("Expected Proto '%s', got '%s'", proto, config.Proto)
		}
	}
}

// TestParseConfigDevices tests different device types
func TestParseConfigDevices(t *testing.T) {
	devices := []string{"tun", "tap"}

	for _, dev := range devices {
		content := "status /var/log/openvpn/status.log\ndev " + dev

		tmpfile := createTempFile(t, "server-dev-*.conf", content)
		config, err := ParseConfig(tmpfile)
		os.Remove(tmpfile)

		if err != nil {
			t.Fatalf("ParseConfig failed for dev %s: %v", dev, err)
		}

		if config.Dev != dev {
			t.Errorf("Expected Dev '%s', got '%s'", dev, config.Dev)
		}
	}
}

// TestParseConfigEmptyLines tests handling of empty lines
func TestParseConfigEmptyLines(t *testing.T) {
	content := `

local 192.168.1.100

port 1194

status /var/log/openvpn/status.log

`

	tmpfile := createTempFile(t, "server-empty-*.conf", content)
	defer os.Remove(tmpfile)

	config, err := ParseConfig(tmpfile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.Local != "192.168.1.100" {
		t.Errorf("Expected Local '192.168.1.100', got '%s'", config.Local)
	}
}

// TestGetServerID tests the server ID generation
func TestGetServerID(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/var/log/openvpn/status.log", "status"},
		{"/var/log/openvpn/server1-status.log", "server1-status"},
		{"/var/log/openvpn/vpn.log", "vpn"},
		{"status.log", "status"},
		{"/path/to/my-server.log", "my-server"},
		{"/path/to/file", "file"},
		{"/path/to/file.txt.log", "file.txt"},
	}

	for _, tt := range tests {
		result := getServerID(tt.path)
		if result != tt.expected {
			t.Errorf("getServerID(%s) = '%s', expected '%s'", tt.path, result, tt.expected)
		}
	}
}

// TestParseConfigWhitespace tests handling of various whitespace
func TestParseConfigWhitespace(t *testing.T) {
	content := "local    192.168.1.100\nport\t1194\nstatus /var/log/openvpn/status.log   30"

	tmpfile := createTempFile(t, "server-whitespace-*.conf", content)
	defer os.Remove(tmpfile)

	config, err := ParseConfig(tmpfile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.Local != "192.168.1.100" {
		t.Errorf("Expected Local '192.168.1.100', got '%s'", config.Local)
	}
	if config.Port != "1194" {
		t.Errorf("Expected Port '1194', got '%s'", config.Port)
	}
}

// TestParseConfigCustomPort tests non-default port values
func TestParseConfigCustomPort(t *testing.T) {
	content := `status /var/log/openvpn/status.log
port 5000`

	tmpfile := createTempFile(t, "server-custom-port-*.conf", content)
	defer os.Remove(tmpfile)

	config, err := ParseConfig(tmpfile)
	if err != nil {
		t.Fatalf("ParseConfig failed: %v", err)
	}

	if config.Port != "5000" {
		t.Errorf("Expected Port '5000', got '%s'", config.Port)
	}
}

// Helper function to create temporary files for testing
func createTempFile(t *testing.T, pattern, content string) string {
	tmpfile, err := os.CreateTemp("", pattern)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	if err := tmpfile.Close(); err != nil {
		os.Remove(tmpfile.Name())
		t.Fatalf("Failed to close temp file: %v", err)
	}

	return tmpfile.Name()
}

// BenchmarkParseConfig benchmarks config file parsing
func BenchmarkParseConfig(b *testing.B) {
	content := `local 192.168.1.100
port 1194
proto udp
dev tun
status /var/log/openvpn/status.log 30
status-version 3`

	tmpfile, _ := os.CreateTemp("", "benchmark-*.conf")
	tmpfile.Write([]byte(content))
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseConfig(tmpfile.Name())
	}
}
