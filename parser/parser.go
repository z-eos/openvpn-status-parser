package parser

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ParseFile reads and parses an OpenVPN status file with specified version.
// It returns the parsed Status and any errors encountered during parsing.
// Parsing continues even if errors occur, collecting all errors for reporting.
//
// Version formats:
// v1: Comma-separated, basic fields only (CommonName, RealAddress, BytesReceived, BytesSent, ConnectedSince)
// v2: Comma-separated, extended fields (adds VirtualAddress, VirtualIPv6Address, Username, ClientID, PeerID, DataCipher)
// v3: Tab-separated, same fields as v2
func ParseFile(filepath string, version StatusVersion) (*Status, []error) {
	// Open the status file
	file, err := os.Open(filepath)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to open file: %w", err)}
	}
	defer file.Close()

	// Determine delimiter based on version
	delimiter := ","
	if version == Version3 {
		delimiter = "\t"
	}

	// Initialize empty status structure
	status := &Status{
		ClientList:   make([]Client, 0),
		RoutingTable: make([]Route, 0),
		Time:         make([]string, 0),
	}

	// Track parsing errors without stopping
	var parseErrors []error
	lineNum := 0

	// Read file line by line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse the line and collect any errors
		if err := parseLine(line, status, lineNum, version, delimiter); err != nil {
			parseErrors = append(parseErrors, err)
		}
	}

	// Check for scanner errors
	if err := scanner.Err(); err != nil {
		parseErrors = append(parseErrors, fmt.Errorf("error reading file: %w", err))
	}

	return status, parseErrors
}

// parseLine processes a single line from the status file.
// It identifies the line type and delegates to appropriate handler.
func parseLine(line string, status *Status, lineNum int, version StatusVersion, delimiter string) error {
	// Split line by delimiter
	fields := strings.Split(line, delimiter)
	if len(fields) == 0 {
		return nil
	}

	// For v1, there are no line type prefixes - all lines are client data
	if version == Version1 {
		return handleClientListV1(fields, status, lineNum)
	}

	// For v2/v3, identify line type by first field
	lineType := fields[0]

	switch lineType {
	case "TITLE":
		return handleTitle(fields, status, lineNum)
	case "TIME":
		return handleTime(fields, status, lineNum)
	case "HEADER":
		// We ignore headers (Option A from discussion)
		return nil
	case "CLIENT_LIST":
		return handleClientListV2V3(fields, status, lineNum)
	case "ROUTING_TABLE":
		return handleRoutingTable(fields, status, lineNum)
	default:
		// Unknown line type - not necessarily an error, might be future extension
		return nil
	}
}

// handleTitle parses TITLE lines (v2/v3 only).
// Format: TITLE<delimiter><server description>
func handleTitle(fields []string, status *Status, lineNum int) error {
	if len(fields) < 2 {
		return ParseError{
			Line:  lineNum,
			Field: "TITLE",
			Value: strings.Join(fields, ","),
			Err:   fmt.Errorf("expected at least 2 fields, got %d", len(fields)),
		}
	}
	status.Title = fields[1]
	return nil
}

// handleTime parses TIME lines (v2/v3 only).
// Format: TIME<delimiter><human readable time><delimiter><epoch time>
func handleTime(fields []string, status *Status, lineNum int) error {
	if len(fields) < 2 {
		return ParseError{
			Line:  lineNum,
			Field: "TIME",
			Value: strings.Join(fields, ","),
			Err:   fmt.Errorf("expected at least 2 fields, got %d", len(fields)),
		}
	}
	// Store all time fields (excluding the "TIME" prefix)
	status.Time = fields[1:]
	return nil
}

// handleClientListV1 parses client lines in v1 format.
// Format: <CommonName>,<RealAddress>,<BytesReceived>,<BytesSent>,<ConnectedSince>
// Example: user1,1.2.3.4:12345,1024000,2048000,Mon Jan 15 10:30:45 2024
func handleClientListV1(fields []string, status *Status, lineNum int) error {
	// v1 CLIENT_LIST should have 5 fields
	expectedFields := 5
	if len(fields) < expectedFields {
		return ParseError{
			Line:  lineNum,
			Field: "CLIENT_LIST_V1",
			Value: strings.Join(fields, ","),
			Err:   fmt.Errorf("expected %d fields, got %d", expectedFields, len(fields)),
		}
	}

	client := Client{}
	var errs []error

	// Parse fields
	client.CommonName = fields[0]
	// real address comes with port
	aparts := strings.Split(fields[1], ":")
	client.RealAddress = aparts[0]

	// Parse numeric fields
	if fields[2] != "" {
		if val, err := strconv.ParseInt(fields[2], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "bytesReceived", Value: fields[2], Err: err})
		} else {
			client.BytesReceived = val
		}
	}

	if fields[3] != "" {
		if val, err := strconv.ParseInt(fields[3], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "bytesSent", Value: fields[3], Err: err})
		} else {
			client.BytesSent = val
		}
	}

	client.ConnectedSince = fields[4]

	// Add client even if some fields had errors
	status.ClientList = append(status.ClientList, client)

	// Return first error if any occurred
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// handleClientListV2V3 parses CLIENT_LIST lines in v2/v3 format.
// Format: CLIENT_LIST<delimiter><commonName><delimiter><realAddress><delimiter>...
// Fields: CommonName, RealAddress, VirtualAddress, VirtualIPv6Address,
//
//	BytesReceived, BytesSent, ConnectedSince, ConnectedSinceTime,
//	Username, ClientID, PeerID, [DataCipher]
func handleClientListV2V3(fields []string, status *Status, lineNum int) error {
	// CLIENT_LIST should have at least 12 fields: prefix + 11 data fields
	// v2/v3 may have additional optional fields like DataCipher
	minFields := 12
	if len(fields) < minFields {
		return ParseError{
			Line:  lineNum,
			Field: "CLIENT_LIST",
			Value: strings.Join(fields, ","),
			Err:   fmt.Errorf("expected at least %d fields, got %d", minFields, len(fields)),
		}
	}

	client := Client{}
	var errs []error

	// Parse each field with error collection
	client.CommonName = fields[1]
	// real address comes with port
	aparts := strings.Split(fields[2], ":")
	client.RealAddress = aparts[0]
	client.VirtualAddress = fields[3]
	client.VirtualIPv6Address = fields[4]

	// Parse numeric fields with error handling
	if fields[5] != "" {
		if val, err := strconv.ParseInt(fields[5], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "bytesReceived", Value: fields[5], Err: err})
		} else {
			client.BytesReceived = val
		}
	}

	if fields[6] != "" {
		if val, err := strconv.ParseInt(fields[6], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "bytesSent", Value: fields[6], Err: err})
		} else {
			client.BytesSent = val
		}
	}

	client.ConnectedSince = fields[7]

	if fields[8] != "" {
		if val, err := strconv.ParseInt(fields[8], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "connectedSinceTime", Value: fields[8], Err: err})
		} else {
			client.ConnectedSinceTime = val
		}
	}

	client.Username = fields[9]

	if fields[10] != "" {
		if val, err := strconv.ParseInt(fields[10], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "clientId", Value: fields[10], Err: err})
		} else {
			client.ClientID = val
		}
	}

	if fields[11] != "" {
		if val, err := strconv.ParseInt(fields[11], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "peerId", Value: fields[11], Err: err})
		} else {
			client.PeerID = val
		}
	}

	// Optional field: DataCipher (field 12, index 12)
	if len(fields) > 12 && fields[12] != "" {
		client.DataCipher = fields[12]
	}

	// Add client even if some fields had errors
	status.ClientList = append(status.ClientList, client)

	// Return first error if any occurred
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

// handleRoutingTable parses ROUTING_TABLE lines (v2/v3 only).
// Format: ROUTING_TABLE<delimiter><virtualAddress><delimiter><commonName><delimiter>...
func handleRoutingTable(fields []string, status *Status, lineNum int) error {
	// ROUTING_TABLE should have 6 fields: prefix + 5 data fields
	expectedFields := 6
	if len(fields) < expectedFields {
		return ParseError{
			Line:  lineNum,
			Field: "ROUTING_TABLE",
			Value: strings.Join(fields, ","),
			Err:   fmt.Errorf("expected %d fields, got %d", expectedFields, len(fields)),
		}
	}

	route := Route{}
	var errs []error

	// Parse string fields
	route.VirtualAddress = fields[1]
	route.CommonName = fields[2]
	// real address comes with port
	aparts := strings.Split(fields[3], ":")
	route.RealAddress = aparts[0]
	route.LastRef = fields[4]

	// Parse numeric field
	if fields[5] != "" {
		if val, err := strconv.ParseInt(fields[5], 10, 64); err != nil {
			errs = append(errs, ParseError{Line: lineNum, Field: "lastRefTime", Value: fields[5], Err: err})
		} else {
			route.LastRefTime = val
		}
	}

	// Add route even if some fields had errors
	status.RoutingTable = append(status.RoutingTable, route)

	// Return first error if any occurred
	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}
