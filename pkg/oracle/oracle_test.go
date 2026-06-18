package oracle

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

const ConnectionFilename = "test_connections.csv"
const ConnectionFilename2 = "test2_connections.csv"

func setup(tb testing.TB) func(tb testing.TB) {
	os.WriteFile(ConnectionFilename, []byte("\n#\nlabel,username/password@database"), 0644)

	// Return a function to teardown the test
	return func(tb testing.TB) {
		os.Remove(ConnectionFilename)
	}
}

func TestParseConnectionsFromFile(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	fileParams := FileParams{
		File: ConnectionFilename}

	connections, err := ParseConnectionsFromFile(fileParams)
	assert.Nil(t, err)
	assert.Equal(t, len(*connections), 1)
	connection := (*connections)[0]
	assert.Equal(t, connection.Label, "label")
	assert.Equal(t, connection.Username, "username")
	assert.Equal(t, connection.Password, "password")
	assert.Equal(t, connection.Database, "database")
}

func TestParseConnectionsFromNoFile(t *testing.T) {
	teardown := setup(t)
	defer teardown(t)

	fileParams := FileParams{
		File: ""}

	connections, err := ParseConnectionsFromFile(fileParams)

	assert.NotNil(t, err)

	noConnections := []Connection{}
	assert.Equal(t, connections, &noConnections)
}

func setupInvalidFile(tb testing.TB) func(tb testing.TB) {
	os.WriteFile(ConnectionFilename2, []byte("blablabla"), 0644)

	// Return a function to teardown the test
	return func(tb testing.TB) {
		os.Remove(ConnectionFilename2)
	}
}

func TestParseConnectionsFromFileInvalidFile(t *testing.T) {
	teardown := setupInvalidFile(t)
	defer teardown(t)

	fileParams := FileParams{
		File: ConnectionFilename2}

	connections, err := ParseConnectionsFromFile(fileParams)

	assert.NotNil(t, err)
	noConnections := []Connection{}
	assert.Equal(t, connections, &noConnections)
}

func TestExtractOracleErrorWhenNil(t *testing.T) {
	err := ExtractOracleError(nil)
	assert.Nil(t, err)
}

func TestExtractOracleErrorWhenNormalError(t *testing.T) {
	anErr := fmt.Errorf("Not a oracle error")
	err := ExtractOracleError(anErr)
	assert.Equal(t, err, anErr)
}

func TestExtractOracleError(t *testing.T) {
	// don't know how to create an godror.OraErr
}
