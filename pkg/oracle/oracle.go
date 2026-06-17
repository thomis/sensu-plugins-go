package oracle

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/godror/godror"
)

type FileParams struct {
	File         string
	Timeout      time.Duration
	ExcludeTypes []string
}

type Connection struct {
	Label        string
	Username     string
	Password     string
	Database     string
	Timeout      time.Duration
	ExcludeTypes []string
}

func ParseConnectionsFromFile(fileParams FileParams) (*[]Connection, error) {
	connections := []Connection{}

	readFile, err := os.Open(fileParams.File)
	if err != nil {
		return &connections, err
	}
	defer readFile.Close()

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	reConnection := regexp.MustCompile(`(.+),(.+)/(.+)@(.+)`)

	i := 0
	for fileScanner.Scan() {
		i++
		line := strings.TrimSpace(fileScanner.Text())

		// empty line
		if len(line) == 0 {
			continue
		}

		// comment line
		if line[0] == '#' {
			continue
		}

		result := reConnection.FindSubmatch([]byte(line))
		if len(result) == 0 {
			return &connections, fmt.Errorf("connection string on line [%d] does not match pattern [label,username/password@database]", i)
		}

		connection := Connection{
			Label:        string(result[1]),
			Username:     string(result[2]),
			Password:     string(result[3]),
			Database:     string(result[4]),
			Timeout:      fileParams.Timeout,
			ExcludeTypes: fileParams.ExcludeTypes}
		connections = append(connections, connection)
	}

	return &connections, nil
}

// IsPLSQL reports whether the given statement is an anonymous PL/SQL block
// (and therefore should be executed with OUT bind variables rather than queried
// as a result set). Detection is based on the leading keyword.
func IsPLSQL(query string) bool {
	q := strings.ToLower(strings.TrimSpace(query))
	return strings.HasPrefix(q, "begin") || strings.HasPrefix(q, "declare")
}

// ReadQuery resolves the query text from an inline string or a file. Exactly one
// of query or file must be provided. A trailing statement terminator is stripped
// (";" for plain SQL, an optional SQL*Plus "/" for PL/SQL blocks) so the
// statement is accepted by the driver.
func ReadQuery(query string, file string) (string, error) {
	hasQuery := len(strings.TrimSpace(query)) > 0
	hasFile := len(strings.TrimSpace(file)) > 0

	switch {
	case hasQuery && hasFile:
		return "", fmt.Errorf("provide either a query (-q) or a query file (--query-file), not both")
	case !hasQuery && !hasFile:
		return "", fmt.Errorf("no query provided (use -q for an inline query or --query-file for a query file)")
	case hasFile:
		data, err := os.ReadFile(file)
		if err != nil {
			return "", err
		}
		query = string(data)
	}

	query = strings.TrimSpace(query)
	if IsPLSQL(query) {
		query = strings.TrimRight(query, " \n\r\t")
		query = strings.TrimSuffix(query, "/")
		query = strings.TrimRight(query, " \n\r\t")
	} else {
		query = strings.TrimRight(query, "; \n\r\t")
	}

	if len(query) == 0 {
		return "", fmt.Errorf("query is empty")
	}

	return query, nil
}

// NormalizeStatus maps a status value returned by a user query to a canonical
// status: "ok", "warning" or "critical". Comparison is case-insensitive and
// ignores surrounding whitespace. An unrecognized value yields an error.
func NormalizeStatus(status string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "ok":
		return "ok", nil
	case "warn", "warning":
		return "warning", nil
	case "error", "critical", "crit":
		return "critical", nil
	default:
		return "", fmt.Errorf("query returned unexpected status %q (expected one of: ok, warn, warning, error)", status)
	}
}

// QueryOutcome is the normalized result of running a query against a single
// connection in batch mode. Status is one of "ok", "warning" or "critical".
type QueryOutcome struct {
	Label   string
	Status  string
	Message string
}

// AggregateQueryOutcomes reduces per-connection outcomes to a single overall
// status (worst-status-wins: critical > warning > ok) and a human-readable
// report. The report starts with a summary line and lists each non-ok
// connection. Any unrecognized status is treated as critical.
func AggregateQueryOutcomes(outcomes []QueryOutcome) (string, string) {
	var ok, warning, critical int
	details := []string{}

	for _, o := range outcomes {
		switch o.Status {
		case "ok":
			ok++
		case "warning":
			warning++
			details = append(details, fmt.Sprintf("- %s: WARNING %s", o.Label, o.Message))
		default:
			critical++
			details = append(details, fmt.Sprintf("- %s: CRITICAL %s", o.Label, o.Message))
		}
	}

	summary := fmt.Sprintf("%d critical, %d warning, %d ok (of %d)", critical, warning, ok, len(outcomes))
	output := summary
	if len(details) > 0 {
		output = summary + "\n" + strings.Join(details, "\n")
	}

	status := "ok"
	switch {
	case critical > 0:
		status = "critical"
	case warning > 0:
		status = "warning"
	}

	return status, output
}

func ExtractOracleError(err error) error {
	if err == nil {
		return err
	}

	oraErr, isOraErr := godror.AsOraErr(err)
	if isOraErr {
		return fmt.Errorf("ORA-%d: %s", oraErr.Code(), oraErr.Message())
	}

	return err
}
