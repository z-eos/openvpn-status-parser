package formatter

import (
	"encoding/json"
	"openvpn-status-parser/parser"
	"strings"
	"testing"
)

// TestJSONFormatterCompact tests compact JSON output
func TestJSONFormatterCompact(t *testing.T) {
	status := createTestStatus()
	formatter := NewJSONFormatter(false)

	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("JSON formatting failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if strings.Contains(output, "\n") {
		t.Error("Compact JSON should not contain newlines")
	}

	if _, ok := result["server"]; !ok {
		t.Error("Expected 'server' field in output")
	}

	if _, ok := result["clientList"]; !ok {
		t.Error("Expected 'clientList' field in output")
	}
}

// TestJSONFormatterIndent tests pretty-printed JSON output
func TestJSONFormatterIndent(t *testing.T) {
	status := createTestStatus()
	formatter := NewJSONFormatter(true)

	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("JSON formatting failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(output), &result); err != nil {
		t.Fatalf("Output is not valid JSON: %v", err)
	}

	if !strings.Contains(output, "\n") {
		t.Error("Indented JSON should contain newlines")
	}

	if !strings.Contains(output, "  ") {
		t.Error("Indented JSON should contain spaces for indentation")
	}
}

// TestJSONFormatterEmptyOptionalFields tests that empty fields are omitted
func TestJSONFormatterEmptyOptionalFields(t *testing.T) {
	status := &parser.Status{
		ClientList: []parser.Client{
			{
				CommonName:    "user1",
				RealAddress:   "192.168.1.100:54321",
				BytesReceived: 1048576,
				BytesSent:     2097152,
			},
		},
	}

	formatter := NewJSONFormatter(false)
	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("JSON formatting failed: %v", err)
	}

	if strings.Contains(output, "virtualAddress") {
		t.Error("Empty virtualAddress should be omitted")
	}
	if strings.Contains(output, "username") {
		t.Error("Empty username should be omitted")
	}
}

// TestOpenMetricsFormatter tests OpenMetrics output format
func TestOpenMetricsFormatter(t *testing.T) {
	status := createTestStatus()
	formatter := NewOpenMetricsFormatter()

	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	if !strings.HasSuffix(strings.TrimSpace(output), "# EOF") {
		t.Error("OpenMetrics output should end with '# EOF'")
	}

	requiredMetrics := []string{
		"openvpn_client_bytes_received_total",
		"openvpn_client_bytes_sent_total",
		"openvpn_client_connected",
		"openvpn_clients_connected_total",
		"openvpn_status_info",
	}

	for _, metric := range requiredMetrics {
		if !strings.Contains(output, metric) {
			t.Errorf("Output should contain metric '%s'", metric)
		}
	}

	if !strings.Contains(output, "# HELP") {
		t.Error("Output should contain HELP comments")
	}
	if !strings.Contains(output, "# TYPE") {
		t.Error("Output should contain TYPE comments")
	}
}

// TestOpenMetricsFormatterServerID tests that server_id label is present
func TestOpenMetricsFormatterServerID(t *testing.T) {
	status := createTestStatus()
	formatter := NewOpenMetricsFormatter()

	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	if !strings.Contains(output, `server_id="test-server"`) {
		t.Error("Output should contain server_id label")
	}
}

// TestOpenMetricsFormatterServerMetadata tests server metadata in info metric
func TestOpenMetricsFormatterServerMetadata(t *testing.T) {
	status := createTestStatus()
	formatter := NewOpenMetricsFormatter()

	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	expectedLabels := []string{
		`local="192.168.1.100"`,
		`port="1194"`,
		`proto="udp"`,
		`dev="tun"`,
	}

	for _, label := range expectedLabels {
		if !strings.Contains(output, label) {
			t.Errorf("Output should contain label '%s'", label)
		}
	}
}

// TestOpenMetricsFormatterLabelEscaping tests special character escaping in labels
func TestOpenMetricsFormatterLabelEscaping(t *testing.T) {
	status := &parser.Status{
		Server: &parser.ServerConfig{
			ID: "test",
		},
		ClientList: []parser.Client{
			{
				CommonName:    `user"with"quotes`,
				RealAddress:   "192.168.1.100:54321",
				BytesReceived: 1000,
				BytesSent:     2000,
			},
		},
	}

	formatter := NewOpenMetricsFormatter()
	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	if !strings.Contains(output, `user\"with\"quotes`) {
		t.Error("Quotes should be escaped in label values")
	}
}

// TestOpenMetricsFormatterNoClients tests output with no clients
func TestOpenMetricsFormatterNoClients(t *testing.T) {
	status := &parser.Status{
		Server: &parser.ServerConfig{
			ID:    "test",
			Local: "192.168.1.100",
			Port:  "1194",
		},
		ClientList: []parser.Client{},
	}

	formatter := NewOpenMetricsFormatter()
	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	if !strings.Contains(output, "openvpn_clients_connected_total") {
		t.Error("Output should contain clients_connected_total metric even with no clients")
	}

	if !strings.Contains(output, "openvpn_clients_connected_total{server_id=\"test\"} 0") {
		t.Error("Should show 0 connected clients")
	}
}

// TestOpenMetricsFormatterRoutingTable tests routing table metrics
func TestOpenMetricsFormatterRoutingTable(t *testing.T) {
	status := createTestStatus()
	formatter := NewOpenMetricsFormatter()

	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	if !strings.Contains(output, "openvpn_routing_entries_total") {
		t.Error("Output should contain routing_entries_total metric")
	}
	if !strings.Contains(output, "openvpn_routing_last_ref_seconds") {
		t.Error("Output should contain routing_last_ref_seconds metric")
	}
}

// TestOpenMetricsFormatterConnectionDuration tests duration calculation
func TestOpenMetricsFormatterConnectionDuration(t *testing.T) {
	status := createTestStatus()
	formatter := NewOpenMetricsFormatter()

	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	if !strings.Contains(output, "openvpn_client_connected_duration_seconds") {
		t.Error("Output should contain connection duration metric")
	}

	if !strings.Contains(output, "openvpn_client_connected_duration_seconds{") {
		t.Error("Duration metric should have labels")
	}
}

// TestOpenMetricsFormatterV1NoTimestamp tests v1 format without timestamps
func TestOpenMetricsFormatterV1NoTimestamp(t *testing.T) {
	status := &parser.Status{
		Server: &parser.ServerConfig{
			ID: "test",
		},
		ClientList: []parser.Client{
			{
				CommonName:         "user1",
				RealAddress:        "192.168.1.100:54321",
				BytesReceived:      1000,
				BytesSent:          2000,
				ConnectedSinceTime: 0,
			},
		},
	}

	formatter := NewOpenMetricsFormatter()
	output, err := formatter.Format(status)
	if err != nil {
		t.Fatalf("OpenMetrics formatting failed: %v", err)
	}

	if strings.Contains(output, "openvpn_client_connected_duration_seconds") {
		t.Error("Should not output duration metric for v1 format without timestamps")
	}
}

// Helper function to create a test status structure
func createTestStatus() *parser.Status {
	return &parser.Status{
		Server: &parser.ServerConfig{
			ID:    "test-server",
			Local: "192.168.1.100",
			Port:  "1194",
			Proto: "udp",
			Dev:   "tun",
		},
		Title: "Test OpenVPN Server",
		Time:  []string{"Thu Nov 27 10:30:45 2025", "1732704645"},
		ClientList: []parser.Client{
			{
				CommonName:         "user1",
				RealAddress:        "192.168.1.100:54321",
				VirtualAddress:     "10.8.0.2",
				BytesReceived:      1048576,
				BytesSent:          2097152,
				ConnectedSince:     "Thu Nov 27 09:30:45 2025",
				ConnectedSinceTime: 1732700645,
				Username:           "user1",
				ClientID:           0,
				PeerID:             0,
				DataCipher:         "AES-256-GCM",
			},
			{
				CommonName:         "alice",
				RealAddress:        "203.0.113.50:12345",
				VirtualAddress:     "10.8.0.6",
				BytesReceived:      5242880,
				BytesSent:          10485760,
				ConnectedSince:     "Thu Nov 27 08:15:30 2025",
				ConnectedSinceTime: 1732696530,
				Username:           "alice",
				ClientID:           1,
				PeerID:             1,
				DataCipher:         "AES-256-GCM",
			},
		},
		RoutingTable: []parser.Route{
			{
				VirtualAddress: "10.8.0.2",
				CommonName:     "user1",
				RealAddress:    "192.168.1.100:54321",
				LastRef:        "Thu Nov 27 10:30:45 2025",
				LastRefTime:    1732704645,
			},
		},
	}
}

// BenchmarkJSONFormatter benchmarks JSON formatting
func BenchmarkJSONFormatter(b *testing.B) {
	status := createTestStatus()
	formatter := NewJSONFormatter(false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.Format(status)
	}
}

// BenchmarkOpenMetricsFormatter benchmarks OpenMetrics formatting
func BenchmarkOpenMetricsFormatter(b *testing.B) {
	status := createTestStatus()
	formatter := NewOpenMetricsFormatter()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		formatter.Format(status)
	}
}
