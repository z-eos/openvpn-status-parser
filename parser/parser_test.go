package parser

import (
	"os"
	"testing"
)

// TestParseFileV1 tests parsing of version 1 status files
func TestParseFileV1(t *testing.T) {
	content := `user1,192.168.1.100:54321,1048576,2097152,Thu Nov 27 09:30:45 2025
alice,203.0.113.50:12345,5242880,10485760,Thu Nov 27 08:15:30 2025
bob,198.51.100.25:33456,15728640,31457280,Wed Nov 26 22:45:00 2025`

	tmpfile := createTempFile(t, "status-v1-*.log", content)
	defer os.Remove(tmpfile)

	status, errors := ParseFile(tmpfile, Version1)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	if len(status.ClientList) != 3 {
		t.Errorf("Expected 3 clients, got %d", len(status.ClientList))
	}

	client := status.ClientList[0]
	if client.CommonName != "user1" {
		t.Errorf("Expected CommonName 'user1', got '%s'", client.CommonName)
	}
	if client.RealAddress != "192.168.1.100:54321" {
		t.Errorf("Expected RealAddress '192.168.1.100:54321', got '%s'", client.RealAddress)
	}
	if client.BytesReceived != 1048576 {
		t.Errorf("Expected BytesReceived 1048576, got %d", client.BytesReceived)
	}
	if client.BytesSent != 2097152 {
		t.Errorf("Expected BytesSent 2097152, got %d", client.BytesSent)
	}

	if len(status.RoutingTable) != 0 {
		t.Errorf("Expected no routing table in v1, got %d entries", len(status.RoutingTable))
	}
}

// TestParseFileV2 tests parsing of version 2 status files (comma-separated)
func TestParseFileV2(t *testing.T) {
	content := `TITLE,OpenVPN Server Status
TIME,Thu Nov 27 10:30:45 2025,1732704645
HEADER,CLIENT_LIST,Common Name,Real Address,Virtual Address,Virtual IPv6 Address,Bytes Received,Bytes Sent,Connected Since,Connected Since (time_t),Username,Client ID,Peer ID,Data Channel Cipher
CLIENT_LIST,user1,192.168.1.100:54321,10.8.0.2,,1048576,2097152,Thu Nov 27 09:30:45 2025,1732700645,user1,0,0,AES-256-GCM
CLIENT_LIST,alice,203.0.113.50:12345,10.8.0.6,,5242880,10485760,Thu Nov 27 08:15:30 2025,1732696530,alice,1,1,AES-256-GCM
HEADER,ROUTING_TABLE,Virtual Address,Common Name,Real Address,Last Ref,Last Ref (time_t)
ROUTING_TABLE,10.8.0.2,user1,192.168.1.100:54321,Thu Nov 27 10:30:45 2025,1732704645
ROUTING_TABLE,10.8.0.6,alice,203.0.113.50:12345,Thu Nov 27 10:30:44 2025,1732704644`

	tmpfile := createTempFile(t, "status-v2-*.log", content)
	defer os.Remove(tmpfile)

	status, errors := ParseFile(tmpfile, Version2)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	if status.Title != "OpenVPN Server Status" {
		t.Errorf("Expected Title 'OpenVPN Server Status', got '%s'", status.Title)
	}

	if len(status.Time) != 2 {
		t.Errorf("Expected 2 time fields, got %d", len(status.Time))
	}

	if len(status.ClientList) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(status.ClientList))
	}

	client := status.ClientList[0]
	if client.VirtualAddress != "10.8.0.2" {
		t.Errorf("Expected VirtualAddress '10.8.0.2', got '%s'", client.VirtualAddress)
	}
	if client.ConnectedSinceTime != 1732700645 {
		t.Errorf("Expected ConnectedSinceTime 1732700645, got %d", client.ConnectedSinceTime)
	}
	if client.Username != "user1" {
		t.Errorf("Expected Username 'user1', got '%s'", client.Username)
	}
	if client.DataCipher != "AES-256-GCM" {
		t.Errorf("Expected DataCipher 'AES-256-GCM', got '%s'", client.DataCipher)
	}

	if len(status.RoutingTable) != 2 {
		t.Errorf("Expected 2 routing entries, got %d", len(status.RoutingTable))
	}

	route := status.RoutingTable[0]
	if route.VirtualAddress != "10.8.0.2" {
		t.Errorf("Expected route VirtualAddress '10.8.0.2', got '%s'", route.VirtualAddress)
	}
	if route.CommonName != "user1" {
		t.Errorf("Expected route CommonName 'user1', got '%s'", route.CommonName)
	}
}

// TestParseFileV3 tests parsing of version 3 status files (tab-separated)
func TestParseFileV3(t *testing.T) {
	content := "TITLE\tOpenVPN Server Status\n" +
		"TIME\tThu Nov 27 10:30:45 2025\t1732704645\n" +
		"HEADER\tCLIENT_LIST\tCommon Name\tReal Address\tVirtual Address\tVirtual IPv6 Address\tBytes Received\tBytes Sent\tConnected Since\tConnected Since (time_t)\tUsername\tClient ID\tPeer ID\tData Channel Cipher\n" +
		"CLIENT_LIST\tuser1\t192.168.1.100:54321\t10.8.0.2\t\t1048576\t2097152\tThu Nov 27 09:30:45 2025\t1732700645\tuser1\t0\t0\tAES-256-GCM\n" +
		"HEADER\tROUTING_TABLE\tVirtual Address\tCommon Name\tReal Address\tLast Ref\tLast Ref (time_t)\n" +
		"ROUTING_TABLE\t10.8.0.2\tuser1\t192.168.1.100:54321\tThu Nov 27 10:30:45 2025\t1732704645\n"

	tmpfile := createTempFile(t, "status-v3-*.log", content)
	defer os.Remove(tmpfile)

	status, errors := ParseFile(tmpfile, Version3)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	if len(status.ClientList) != 1 {
		t.Errorf("Expected 1 client, got %d", len(status.ClientList))
	}

	client := status.ClientList[0]
	if client.CommonName != "user1" {
		t.Errorf("Expected CommonName 'user1', got '%s'", client.CommonName)
	}
}

// TestParseFileEmpty tests parsing of empty file
func TestParseFileEmpty(t *testing.T) {
	tmpfile := createTempFile(t, "status-empty-*.log", "")
	defer os.Remove(tmpfile)

	status, errors := ParseFile(tmpfile, Version3)

	if len(errors) > 0 {
		t.Errorf("Expected no errors for empty file, got %d", len(errors))
	}

	if len(status.ClientList) != 0 {
		t.Errorf("Expected 0 clients for empty file, got %d", len(status.ClientList))
	}
}

// TestParseFileInvalidPath tests handling of non-existent file
func TestParseFileInvalidPath(t *testing.T) {
	_, errors := ParseFile("/nonexistent/path/status.log", Version3)

	if len(errors) == 0 {
		t.Error("Expected error for non-existent file, got none")
	}
}

// TestParseFileMalformed tests handling of malformed data
func TestParseFileMalformed(t *testing.T) {
	content := `TITLE	OpenVPN Server
CLIENT_LIST	user1	192.168.1.100:54321	10.8.0.2		invalid_number	2097152	Thu Nov 27 09:30:45 2025	1732700645	user1	0	0`

	tmpfile := createTempFile(t, "status-malformed-*.log", content)
	defer os.Remove(tmpfile)

	status, errors := ParseFile(tmpfile, Version3)

	if len(errors) == 0 {
		t.Error("Expected parsing errors for malformed data, got none")
	}

	if len(status.ClientList) != 1 {
		t.Errorf("Expected 1 client despite errors, got %d", len(status.ClientList))
	}
}

// TestParseFileV1InsufficientFields tests v1 with missing fields
func TestParseFileV1InsufficientFields(t *testing.T) {
	content := `user1,192.168.1.100:54321,1048576`

	tmpfile := createTempFile(t, "status-v1-bad-*.log", content)
	defer os.Remove(tmpfile)

	_, errors := ParseFile(tmpfile, Version1)

	if len(errors) == 0 {
		t.Error("Expected error for insufficient fields, got none")
	}
}

// TestParseFileWithComments tests that empty lines are ignored (v3 doesn't use # comments)
func TestParseFileWithEmptyLines(t *testing.T) {
	content := `TITLE	OpenVPN Server

CLIENT_LIST	user1	192.168.1.100:54321	10.8.0.2		1048576	2097152	Thu Nov 27 09:30:45 2025	1732700645	user1	0	0	AES-256-GCM

`

	tmpfile := createTempFile(t, "status-empty-lines-*.log", content)
	defer os.Remove(tmpfile)

	status, errors := ParseFile(tmpfile, Version3)

	if len(errors) > 0 {
		t.Errorf("Expected no errors with empty lines, got %d: %v", len(errors), errors)
	}

	if len(status.ClientList) != 1 {
		t.Errorf("Expected 1 client, got %d", len(status.ClientList))
	}
}

// TestParseFileIPv6 tests parsing of IPv6 addresses
func TestParseFileIPv6(t *testing.T) {
	content := "CLIENT_LIST\tbob\t198.51.100.25:33456\t10.8.0.10\tfd00::10\t15728640\t31457280\tWed Nov 26 22:45:00 2025\t1732662300\tbob\t2\t2\tAES-256-GCM\n"

	tmpfile := createTempFile(t, "status-ipv6-*.log", content)
	defer os.Remove(tmpfile)

	status, errors := ParseFile(tmpfile, Version3)

	if len(errors) > 0 {
		t.Errorf("Expected no errors, got %d: %v", len(errors), errors)
	}

	if len(status.ClientList) != 1 {
		t.Fatalf("Expected 1 client, got %d", len(status.ClientList))
	}

	client := status.ClientList[0]
	if client.VirtualIPv6Address != "fd00::10" {
		t.Errorf("Expected IPv6 'fd00::10', got '%s'", client.VirtualIPv6Address)
	}
}

// TestParseErrorType tests the ParseError type
func TestParseErrorType(t *testing.T) {
	err := ParseError{
		Line:  10,
		Field: "bytesReceived",
		Value: "invalid",
		Err:   os.ErrInvalid,
	}

	expected := `line 10, field bytesReceived, value "invalid": invalid argument`
	if err.Error() != expected {
		t.Errorf("Expected error message '%s', got '%s'", expected, err.Error())
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

// BenchmarkParseFileV3Small benchmarks parsing of small status file
func BenchmarkParseFileV3Small(b *testing.B) {
	content := "TITLE\tOpenVPN Server\n" +
		"CLIENT_LIST\tuser1\t192.168.1.100:54321\t10.8.0.2\t\t1048576\t2097152\tThu Nov 27 09:30:45 2025\t1732700645\tuser1\t0\t0\tAES-256-GCM\n"

	tmpfile, _ := os.CreateTemp("", "benchmark-*.log")
	tmpfile.Write([]byte(content))
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseFile(tmpfile.Name(), Version3)
	}
}

// BenchmarkParseFileV3Large benchmarks parsing of large status file
func BenchmarkParseFileV3Large(b *testing.B) {
	content := "TITLE\tOpenVPN Server\n"
	for i := 0; i < 100; i++ {
		content += "CLIENT_LIST\tuser\t192.168.1.100:54321\t10.8.0.2\t\t1048576\t2097152\tThu Nov 27 09:30:45 2025\t1732700645\tuser1\t0\t0\tAES-256-GCM\n"
	}

	tmpfile, _ := os.CreateTemp("", "benchmark-large-*.log")
	tmpfile.Write([]byte(content))
	tmpfile.Close()
	defer os.Remove(tmpfile.Name())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseFile(tmpfile.Name(), Version3)
	}
}
