package parser

import "fmt"

// StatusVersion represents the OpenVPN status file version
type StatusVersion int

const (
	// Version1 - Traditional format with comma-separated basic fields
	Version1 StatusVersion = 1
	// Version2 - Extended format with comma-separated fields including virtual IPs, username, etc.
	Version2 StatusVersion = 2
	// Version3 - Same as v2 but tab-separated instead of comma-separated
	Version3 StatusVersion = 3
)

// Status represents the complete OpenVPN status file structure.
// This matches the v2/v3 format output structure.
type Status struct {
	Server *ServerConfig `json:"server,omitempty"`

	// Title contains the OpenVPN server title/description (v2/v3 only)
	Title string `json:"title,omitempty"`

	// Time contains timestamp information from the status file (v2/v3 only)
	// Usually includes human-readable time and epoch time
	Time []string `json:"time,omitempty"`

	// ClientList contains all connected clients
	ClientList []Client `json:"clientList"`

	// RoutingTable contains virtual IP to client mappings (v2/v3 only)
	RoutingTable []Route `json:"routingTable,omitempty"`
}

type ServerConfig struct {
	// server ID is a basename of status file
	ID string `json:"id"`

	// openvpn options, see openvpn(8)

	//  --local
	Local string `json:"local,omitempty"`

	// --port
	Port string `json:"port,omitempty"`

	// --proto
	Proto string `json:"proto,omitempty"`

	// --dev
	Dev string `json:"dev,omitempty"`
}

// Client represents a single connected OpenVPN client.
// Fields availability depends on status file version:
// - v1: CommonName, RealAddress, BytesReceived, BytesSent, ConnectedSince
// - v2/v3: All fields
type Client struct {
	// CommonName is the client's certificate common name (CN) - all versions
	CommonName string `json:"commonName"`

	// RealAddress is the client's actual IP:port (e.g., "1.2.3.4:12345") - all versions
	RealAddress string `json:"realAddress"`

	// VirtualAddress is the assigned VPN IP (e.g., "10.8.0.2") - v2/v3 only
	VirtualAddress string `json:"virtualAddress,omitempty"`

	// VirtualIPv6Address is the assigned IPv6 address (optional) - v2/v3 only
	VirtualIPv6Address string `json:"virtualIPv6Address,omitempty"`

	// BytesReceived is total bytes received from this client - all versions
	BytesReceived int64 `json:"bytesReceived"`

	// BytesSent is total bytes sent to this client - all versions
	BytesSent int64 `json:"bytesSent"`

	// ConnectedSince is human-readable connection start time - all versions
	ConnectedSince string `json:"connectedSince"`

	// ConnectedSinceTime is Unix timestamp when client connected - v2/v3 only
	ConnectedSinceTime int64 `json:"connectedSinceTime,omitempty"`

	// Username is the authenticated username (optional) - v2/v3 only
	Username string `json:"username,omitempty"`

	// ClientID is the internal OpenVPN client identifier - v2/v3 only
	ClientID int64 `json:"clientId,omitempty"`

	// PeerID is the internal peer identifier - v2/v3 only
	PeerID int64 `json:"peerId,omitempty"`

	// DataCipher is the data channel cipher - v2/v3 only (optional field)
	DataCipher string `json:"dataCipher,omitempty"`
}

// Route represents a single routing table entry.
// Only available in v2/v3 formats.
type Route struct {
	// VirtualAddress is the routed VPN IP or network
	VirtualAddress string `json:"virtualAddress"`

	// CommonName is the client certificate CN this route points to
	CommonName string `json:"commonName"`

	// RealAddress is the client's actual IP:port
	RealAddress string `json:"realAddress"`

	// LastRef is human-readable time of last routing table update
	LastRef string `json:"lastRef"`

	// LastRefTime is Unix timestamp of last routing table update
	LastRefTime int64 `json:"lastRefTime"`
}

// ParseError represents an error encountered during parsing.
// We collect these instead of failing on first error.
type ParseError struct {
	// Line is the line number where error occurred (1-indexed)
	Line int

	// Field is the field name that caused the error
	Field string

	// Value is the problematic value we tried to parse
	Value string

	// Err is the underlying error
	Err error
}

// Error implements the error interface for ParseError
func (e ParseError) Error() string {
	return fmt.Sprintf("line %d, field %s, value %q: %v", e.Line, e.Field, e.Value, e.Err)
}
