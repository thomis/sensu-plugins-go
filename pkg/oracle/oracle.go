package oracle

import (
	"os"
	"time"
	"bufio"
	"regexp"
	"fmt"
	"strings"

	"github.com/godror/godror"
)

type FileParams struct {
	File         string
	Timeout      time.Duration
	ExcludeTypes []string
}

type Connection struct {
	Label    string
	Username string
	Password string
	Database string
	Timeout  time.Duration
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
			Label:    string(result[1]),
			Username: string(result[2]),
			Password: string(result[3]),
			Database: string(result[4]),
			Timeout:  fileParams.Timeout,
			ExcludeTypes: fileParams.ExcludeTypes}
		connections = append(connections, connection)
	}

	return &connections, nil
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
