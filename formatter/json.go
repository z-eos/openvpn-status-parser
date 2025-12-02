package formatter

import (
	"encoding/json"
	"openvpn-status-parser/parser"
)

// JSONFormatter formats the status as JSON.
// It uses Go's standard encoding/json package with proper struct tags.
type JSONFormatter struct {
	// Indent controls whether to pretty-print JSON with indentation.
	// If true, uses 2-space indentation. If false, outputs compact JSON.
	Indent bool
}

// NewJSONFormatter creates a new JSON formatter.
func NewJSONFormatter(indent bool) *JSONFormatter {
	return &JSONFormatter{Indent: indent}
}

// Format converts the Status to JSON format.
// Empty optional fields are omitted due to the "omitempty" JSON tags.
func (f *JSONFormatter) Format(status *parser.Status) (string, error) {
	var output []byte
	var err error

	if f.Indent {
		// Pretty-printed JSON with 2-space indentation
		output, err = json.MarshalIndent(status, "", "  ")
	} else {
		// Compact JSON
		output, err = json.Marshal(status)
	}

	if err != nil {
		return "", err
	}

	return string(output), nil
}
