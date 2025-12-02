# openvpn-status-parser

Go application that parses OpenVPN status files and exports data in JSON or OpenMetrics (Prometheus) format. Supports all three OpenVPN status file versions (v1, v2, v3) and automatically extracts configuration data necessary from OpenVPN server config files.

## Features

- **Multi-version support** - Parse OpenVPN status file versions 1, 2, and 3
- **Config file parsing** - Automatically extract status file path and version from OpenVPN config
- **Dual output formats** - JSON for general use, OpenMetrics for Prometheus
- **Server metadata** - Extract and export server configuration (IP, port, protocol, device)
- **Multi-server support** - Handle multiple OpenVPN servers with unique identifiers
- **Error resilient** - Continues parsing on errors, reports issues without failing
- **Zero dependencies** - Uses only Go standard library

---

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/openvpn-status-parser.git
cd openvpn-status-parser

# Build
go build -o openvpn-status-parser

# Install (optional)
sudo cp openvpn-status-parser /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/yourusername/openvpn-status-parser@latest
```

---

## Quick Start

### 1. Configure OpenVPN to generate status file

Add to your OpenVPN server config:

```
status /var/log/openvpn/status.log
status-version 3
```

### 2. Parse status file

```bash
openvpn-status-parser -file /etc/openvpn/server.conf
```

### 3. Export to Prometheus

```bash
# Generate OpenMetrics output
openvpn-status-parser -file /etc/openvpn/server.conf -format openmetrics > /var/lib/node_exporter/textfile_collector/openvpn.prom
```

---

## Usage

Parse OpenVPN server configuration to automatically detect status file location, version, and extract server metadata:

```bash
openvpn-status-parser -file /etc/openvpn/server.conf
```

**Required OpenVPN config directives:**
```
status /var/log/openvpn/status.log    # Required
status-version 3                      # Optional (defaults to 3)
local 192.168.1.100                   # Optional
port 1194                             # Optional (defaults to 1194)
proto udp                             # Optional
dev tun                               # Optional
```

### Command-Line Options

```
-file string
	Path to OpenVPN config file (required)

-format string
	Output format: json or openmetrics (default: json)

-indent
	Pretty-print JSON output (only applies to JSON format)

-version
	Show version information
```

### Examples

```bash
# JSON output from config file
openvpn-status-parser -file /etc/openvpn/server.conf

# Pretty JSON
openvpn-status-parser -file /etc/openvpn/server.conf -format json -indent

# OpenMetrics from config file
openvpn-status-parser -file /etc/openvpn/server.conf -format openmetrics

# Show version
openvpn-status-parser -version
```

---

## Output Formats

### JSON Format

Clean, structured JSON output suitable for APIs, scripts, and general processing.

**Example:**
```json
{
  "server": {
	"id": "status",
	"local": "192.168.1.100",
	"port": "1194",
	"proto": "udp",
	"dev": "tun"
  },
  "title": "OpenVPN Server Status",
  "time": [
	"Thu Nov 27 10:30:45 2025",
	"1732704645"
  ],
  "clientList": [
	{
	  "commonName": "user1",
	  "realAddress": "192.168.1.100:54321",
	  "virtualAddress": "10.8.0.2",
	  "bytesReceived": 1048576,
	  "bytesSent": 2097152,
	  "connectedSince": "Thu Nov 27 09:30:45 2025",
	  "connectedSinceTime": 1732700645,
	  "username": "user1",
	  "clientId": 0,
	  "peerId": 0
	}
  ],
  "routingTable": [
	{
	  "virtualAddress": "10.8.0.2",
	  "commonName": "user1",
	  "realAddress": "192.168.1.100:54321",
	  "lastRef": "Thu Nov 27 10:30:45 2025",
	  "lastRefTime": 1732704645
	}
  ]
}
```

**Features:**
- Empty optional fields are omitted (`omitempty`)
- Pretty-print with `-indent` flag
- Compatible with jq and other JSON tools

### OpenMetrics Format

Prometheus-compatible exposition format for monitoring and alerting.

**Metrics Exported:**

#### Client Metrics (per client)

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `openvpn_client_bytes_received_total` | counter | Total bytes received from client | `server_id`, `common_name`, `real_address`, `virtual_address`, `username`, `cipher` |
| `openvpn_client_bytes_sent_total` | counter | Total bytes sent to client | Same as above |
| `openvpn_client_connected_duration_seconds` | gauge | Time in seconds since client connected | Same as above |
| `openvpn_client_connected` | gauge | Client connection status (always 1) | Same as above |

#### Server-wide Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `openvpn_clients_connected_total` | gauge | Total number of connected clients | `server_id` |
| `openvpn_routing_entries_total` | gauge | Total routing table entries | `server_id` |

#### Routing Metrics

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `openvpn_routing_last_ref_seconds` | gauge | Unix timestamp of last routing table reference | `server_id`, `virtual_address`, `common_name`, `real_address` |

#### Info Metric

| Metric | Type | Description | Labels |
|--------|------|-------------|--------|
| `openvpn_status_info` | gauge | Server metadata (always 1) | `server_id`, `local`, `port`, `proto`, `dev`, `title`, `updated_at` |

**Example Output:**
```
# HELP openvpn_client_bytes_received_total Total bytes received from client
# TYPE openvpn_client_bytes_received_total counter
openvpn_client_bytes_received_total{server_id="status",common_name="user1",real_address="192.168.1.100:54321",virtual_address="10.8.0.2",username="user1"} 1048576

# HELP openvpn_client_bytes_sent_total Total bytes sent to client
# TYPE openvpn_client_bytes_sent_total counter
openvpn_client_bytes_sent_total{server_id="status",common_name="user1",real_address="192.168.1.100:54321",virtual_address="10.8.0.2",username="user1"} 2097152

# HELP openvpn_client_connected_duration_seconds Time in seconds since client connected
# TYPE openvpn_client_connected_duration_seconds gauge
openvpn_client_connected_duration_seconds{server_id="status",common_name="user1",real_address="192.168.1.100:54321",virtual_address="10.8.0.2",username="user1"} 3600

# HELP openvpn_clients_connected_total Total number of connected clients
# TYPE openvpn_clients_connected_total gauge
openvpn_clients_connected_total{server_id="status"} 3

# HELP openvpn_status_info OpenVPN status file and server metadata
# TYPE openvpn_status_info gauge
openvpn_status_info{server_id="status",local="192.168.1.100",port="1194",proto="udp",dev="tun",title="OpenVPN Server Status",updated_at="Thu Nov 27 10:30:45 2025"} 1
# EOF
```

**Grafana dashboard**

File `openvpn-status-parser.json` contains a grafana dashboard sample.

---

## Prometheus Integration

### Node Exporter Textfile Collector

**Setup:**

1. Enable textfile collector in node_exporter:
```bash
# /etc/systemd/system/node_exporter.service
ExecStart=/usr/local/bin/node_exporter --collector.textfile.directory=/var/lib/node_exporter/textfile_collector
```

2. Create collection directory:
```bash
sudo mkdir -p /var/lib/node_exporter/textfile_collector
```

3. Add cron job to update metrics:
```bash
# /etc/cron.d/openvpn-metrics
*/5 * * * * root /usr/local/bin/openvpn-status-parser -file /etc/openvpn/server.conf -format openmetrics > /var/lib/node_exporter/textfile_collector/openvpn.prom.tmp && mv /var/lib/node_exporter/textfile_collector/openvpn.prom.tmp /var/lib/node_exporter/textfile_collector/openvpn.prom
```

4. Configure Prometheus to scrape node_exporter:
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'node'
    static_configs:
      - targets: ['localhost:9100']
```

**Atomic writes:** The `mv` operation ensures Prometheus never reads partial files.

---

## Building from Source

### Prerequisites

- Go 1.21 or later
- Git (for cloning)

### Build Steps

```bash
# Clone repository
git clone https://github.com/.../openvpn-status-parser.git
cd openvpn-status-parser

# Download dependencies (if any)
go mod tidy

# Build for current platform
go build -o openvpn-status-parser

```

### Testing

```bash
# Run with sample files
./openvpn-status-parser -file server.conf.sample

# Run unit tests
go test ./...

# Run with race detector
go test -race ./...

# Benchmark
go test -bench=. ./...
```

---

## Troubleshooting

### Common Issues

#### 1. "no 'status' directive found in config file"

**Cause:** OpenVPN config doesn't have `status` directive

**Solution:** Add to your OpenVPN config:
```
status /var/log/openvpn/status.log
status-version 3
```

#### 2. "failed to open file: permission denied"

**Cause:** Insufficient permissions to read status file

**Solution:**
```bash
# Check permissions
ls -la /var/log/openvpn/status.log

# Fix permissions
sudo chmod 644 /var/log/openvpn/status.log

# Or run parser as root (in cron)
sudo openvpn-status-parser -file /etc/openvpn/server.conf
```

#### 3. "expected X fields, got Y"

**Cause:** Status file version mismatch

**Solution:**
- Check `status-version` in OpenVPN config
- Ensure it matches actual file format
- Restart OpenVPN after config changes

### Validation

```bash
# Validate OpenMetrics output format
openvpn-status-parser -file /etc/openvpn/server.conf -format openmetrics | promtool check metrics

# Validate JSON output
openvpn-status-parser -file /etc/openvpn/server.conf -format json | jq .

# Check for specific metric
openvpn-status-parser -file /etc/openvpn/server.conf -format openmetrics | grep openvpn_clients_connected_total
```
