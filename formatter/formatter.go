package formatter

import "openvpn-status-parser/parser"

// Formatter is the interface for different output formats.
// Implementations can output the parsed status in various formats
// like JSON, OpenMetrics, XML, etc.
type Formatter interface {
	// Format takes a parsed Status and returns the formatted output as a string.
	// Returns an error if formatting fails.
	Format(status *parser.Status) (string, error)
}
