package formatter

import (
	"fmt"
	"openvpn-status-parser/parser"
	"strings"
	"time"
)

// OpenMetricsFormatter formats the status as OpenMetrics/Prometheus exposition format.
// See: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md
type OpenMetricsFormatter struct{}

// NewOpenMetricsFormatter creates a new OpenMetrics formatter.
func NewOpenMetricsFormatter() *OpenMetricsFormatter {
	return &OpenMetricsFormatter{}
}

// Format converts the Status to OpenMetrics format.
// Generates metrics for:
// - Client bytes sent/received (counters)
// - Client connection duration (gauge)
// - Client connected status (gauge, always 1)
// - Total clients/routes (gauges)
// - Routing last reference time (gauge)
// - Status info (info metric)
func (f *OpenMetricsFormatter) Format(status *parser.Status) (string, error) {
	var sb strings.Builder

	// Current time for duration calculations
	now := time.Now().Unix()

	server := *status.Server

	// Write metric metadata and values

	// 1. Client bytes received (counter)
	sb.WriteString("# HELP openvpn_client_bytes_received_total Total bytes received from client\n")
	sb.WriteString("# TYPE openvpn_client_bytes_received_total counter\n")
	for _, client := range status.ClientList {
		labels := f.buildClientLabels(client, server)
		sb.WriteString(fmt.Sprintf("openvpn_client_bytes_received_total%s %d\n", labels, client.BytesReceived))
	}

	// 2. Client bytes sent (counter)
	sb.WriteString("# HELP openvpn_client_bytes_sent_total Total bytes sent to client\n")
	sb.WriteString("# TYPE openvpn_client_bytes_sent_total counter\n")
	for _, client := range status.ClientList {
		labels := f.buildClientLabels(client, server)
		sb.WriteString(fmt.Sprintf("openvpn_client_bytes_sent_total%s %d\n", labels, client.BytesSent))
	}

	// 3. Client connection duration (gauge)
	sb.WriteString("# HELP openvpn_client_connected_duration_seconds Time in seconds since client connected\n")
	sb.WriteString("# TYPE openvpn_client_connected_duration_seconds gauge\n")
	for _, client := range status.ClientList {
		labels := f.buildClientLabels(client, server)
		duration := now - client.ConnectedSinceTime
		sb.WriteString(fmt.Sprintf("openvpn_client_connected_duration_seconds%s %d\n", labels, duration))
	}

	// 4. Client connected indicator (gauge, always 1 since they're in the status file)
	sb.WriteString("# HELP openvpn_client_connected Client connection status (1 = connected)\n")
	sb.WriteString("# TYPE openvpn_client_connected gauge\n")
	for _, client := range status.ClientList {
		labels := f.buildClientLabels(client, server)
		sb.WriteString(fmt.Sprintf("openvpn_client_connected%s 1\n", labels))
	}

	labels := []string{
		fmt.Sprintf("server_id=%q", f.sanitizeLabelValue(server.ID)),
	}

	// 5. Total connected clients (gauge)
	sb.WriteString("# HELP openvpn_clients_connected_total Total number of connected clients\n")
	sb.WriteString("# TYPE openvpn_clients_connected_total gauge\n")
	sb.WriteString(fmt.Sprintf("openvpn_clients_connected_total{%s} %d\n", strings.Join(labels, ","), len(status.ClientList)))

	// 6. Total routing entries (gauge)
	sb.WriteString("# HELP openvpn_routing_entries_total Total number of routing table entries\n")
	sb.WriteString("# TYPE openvpn_routing_entries_total gauge\n")
	sb.WriteString(fmt.Sprintf("openvpn_routing_entries_total{%s} %d\n", strings.Join(labels, ","), len(status.RoutingTable)))

	// 7. Routing table last reference time (gauge)
	sb.WriteString("# HELP openvpn_routing_last_ref_seconds Unix timestamp of last routing table reference\n")
	sb.WriteString("# TYPE openvpn_routing_last_ref_seconds gauge\n")
	for _, route := range status.RoutingTable {
		labels := f.buildRouteLabels(route, server)
		sb.WriteString(fmt.Sprintf("openvpn_routing_last_ref_seconds%s %d\n", labels, route.LastRefTime))
	}

	// 8. Status info metric (info type - gauge with value 1)
	sb.WriteString("# HELP openvpn_status_info OpenVPN status file metadata\n")
	sb.WriteString("# TYPE openvpn_status_info gauge\n")
	infoLabels := f.buildInfoLabels(status, server)
	sb.WriteString(fmt.Sprintf("openvpn_status_info%s 1\n", infoLabels))

	// 9. End of metrics marker (required by OpenMetrics spec)
	sb.WriteString("# EOF\n")

	return sb.String(), nil
}

// buildClientLabels creates label string for client metrics.
// Format: {common_name="...",real_address="...",virtual_address="...",username="..."}
// Empty optional labels (username) are omitted.
func (f *OpenMetricsFormatter) buildClientLabels(client parser.Client, server parser.ServerConfig) string {
	labels := []string{
		fmt.Sprintf("common_name=%q", f.sanitizeLabelValue(client.CommonName)),
		fmt.Sprintf("real_address=%q", f.sanitizeLabelValue(client.RealAddress)),
		fmt.Sprintf("server_id=%q", f.sanitizeLabelValue(server.ID)),
		fmt.Sprintf("virtual_address=%q", f.sanitizeLabelValue(client.VirtualAddress)),
	}

	// Add username only if present
	if client.Username != "" {
		labels = append(labels, fmt.Sprintf("username=%q", f.sanitizeLabelValue(client.Username)))
	}

	return "{" + strings.Join(labels, ",") + "}"
}

// buildRouteLabels creates label string for routing metrics.
// Format: {virtual_address="...",common_name="...",real_address="..."}
func (f *OpenMetricsFormatter) buildRouteLabels(route parser.Route, server parser.ServerConfig) string {
	labels := []string{
		fmt.Sprintf("virtual_address=%q", f.sanitizeLabelValue(route.VirtualAddress)),
		fmt.Sprintf("common_name=%q", f.sanitizeLabelValue(route.CommonName)),
		fmt.Sprintf("real_address=%q", f.sanitizeLabelValue(route.RealAddress)),
		fmt.Sprintf("server_id=%q", f.sanitizeLabelValue(server.ID)),
	}
	return "{" + strings.Join(labels, ",") + "}"
}

// buildInfoLabels creates label string for the info metric.
// Format: {title="...",updated_at="..."}
func (f *OpenMetricsFormatter) buildInfoLabels(status *parser.Status, server parser.ServerConfig) string {
	labels := []string{
		fmt.Sprintf("title=%q", f.sanitizeLabelValue(status.Title)),
		fmt.Sprintf("server_id=%q", f.sanitizeLabelValue(server.ID)),
		fmt.Sprintf("server_local=%q", f.sanitizeLabelValue(server.Local)),
		fmt.Sprintf("server_port=%q", f.sanitizeLabelValue(server.Port)),
		fmt.Sprintf("server_proto=%q", f.sanitizeLabelValue(server.Proto)),
		fmt.Sprintf("server_dev=%q", f.sanitizeLabelValue(server.Dev)),
	}

	// Add timestamp if available
	if len(status.Time) > 0 {
		labels = append(labels, fmt.Sprintf("updated_at=%q", f.sanitizeLabelValue(status.Time[0])))
	}

	return "{" + strings.Join(labels, ",") + "}"
}

// sanitizeLabelValue escapes special characters in label values.
// OpenMetrics requires escaping backslashes, newlines, and double quotes.
// See: https://github.com/OpenObservability/OpenMetrics/blob/main/specification/OpenMetrics.md#escaping
func (f *OpenMetricsFormatter) sanitizeLabelValue(value string) string {
	// Replace backslash first to avoid double-escaping
	value = strings.ReplaceAll(value, "\\", "\\\\")
	// Escape newlines
	value = strings.ReplaceAll(value, "\n", "\\n")
	// Escape double quotes
	value = strings.ReplaceAll(value, "\"", "\\\"")
	return value
}
