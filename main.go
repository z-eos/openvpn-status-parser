package main

import (
	"flag"
	"fmt"
	"openvpn-status-parser/config"
	"openvpn-status-parser/formatter"
	"openvpn-status-parser/parser"
	"os"
	"path/filepath"
)

const (
	Version = "0.1.0"
)

func main() {
	// Define command-line flags
	filePath := flag.String("file", "", "Path to OpenVPN config file (required)")
	format := flag.String("format", "json", "Output format: json or openmetrics")
	indent := flag.Bool("indent", false, "Pretty-print JSON output (only for json format)")
	version := flag.Bool("version", false, "Show version information")

	// Custom usage message
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "OpenVPN Status Parser - Converts OpenVPN status files to JSON or OpenMetrics format\n\n")
		fmt.Fprintf(os.Stderr, "Usage: %s [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -file /etc/openvpn/server.conf\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -file /etc/openvpn/server.conf -format openmetrics\n", os.Args[0])
	}

	flag.Parse()

	// Handle version flag
	if *version {
		fmt.Printf("openvpn-status-parser version %s\n", Version)
		os.Exit(0)
	}

	// Validate required file flag
	if *filePath == "" {
		fmt.Fprintf(os.Stderr, "Error: -file flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate format flag
	if *format != "json" && *format != "openmetrics" {
		fmt.Fprintf(os.Stderr, "Error: -format must be 'json' or 'openmetrics'\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Determine if input is a config file or status file
	var statusFilePath string
	var statusVer parser.StatusVersion
	var serverConfig *parser.ServerConfig

	// Parse OpenVPN config file
	cfg, err := config.ParseConfig(*filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse config file: %v\n", err)
		os.Exit(1)
	}

	// Extract status file path and version from config
	statusFilePath = cfg.StatusFile
	statusVer = getStatusVersion(cfg.StatusVersion)

	// Convert config.ServerConfig to parser.ServerConfig
	serverConfig = &parser.ServerConfig{
		ID:    cfg.ID,
		Local: cfg.Local,
		Port:  cfg.Port,
		Proto: cfg.Proto,
		Dev:   cfg.Dev,
	}

	fmt.Fprintf(os.Stderr, "Config file parsed: server_id=%s, status=%s, version=%d\n",
		serverConfig.ID, statusFilePath, cfg.StatusVersion)

	// Parse the status file
	status, parseErrors := parser.ParseFile(statusFilePath, statusVer)

	// Report any parsing errors to stderr
	if len(parseErrors) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: encountered %d error(s) during parsing:\n", len(parseErrors))
		for _, err := range parseErrors {
			fmt.Fprintf(os.Stderr, "  %v\n", err)
		}
		fmt.Fprintf(os.Stderr, "\n")
	}

	// If status is nil, parsing failed completely
	if status == nil {
		fmt.Fprintf(os.Stderr, "Error: failed to parse status file\n")
		os.Exit(1)
	}

	// Attach server config to status
	status.Server = serverConfig

	// Select formatter based on format flag
	var f formatter.Formatter
	switch *format {
	case "json":
		f = formatter.NewJSONFormatter(*indent)
	case "openmetrics":
		f = formatter.NewOpenMetricsFormatter()
	}

	// Format the output
	output, err := f.Format(status)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to format output: %v\n", err)
		os.Exit(1)
	}

	// Write output to stdout
	fmt.Print(output)

	// Exit with non-zero code if there were parse errors
	if len(parseErrors) > 0 {
		os.Exit(2)
	}
}

// getStatusVersion converts an integer to StatusVersion type
func getStatusVersion(ver int) parser.StatusVersion {
	switch ver {
	case 1:
		return parser.Version1
	case 2:
		return parser.Version2
	case 3:
		return parser.Version3
	default:
		return parser.Version3 // Default to v3
	}
}

// getServerID extracts a server identifier from the status file path.
// Returns the basename without extension.
func getServerID(statusPath string) string {
	basename := filepath.Base(statusPath)
	ext := filepath.Ext(basename)
	if ext != "" {
		basename = basename[:len(basename)-len(ext)]
	}
	return basename
}
