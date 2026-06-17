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

func TestIsPLSQL(t *testing.T) {
	assert.True(t, IsPLSQL("begin null; end;"))
	assert.True(t, IsPLSQL("  BEGIN my_proc(:s, :m); END;"))
	assert.True(t, IsPLSQL("declare x number; begin null; end;"))
	assert.False(t, IsPLSQL("select 'ok', 'fine' from dual"))
	assert.False(t, IsPLSQL("  SELECT 1 FROM dual"))
}

func TestNormalizeStatus(t *testing.T) {
	cases := map[string]string{
		"ok":       "ok",
		"OK":       "ok",
		"  Ok  ":   "ok",
		"warn":     "warning",
		"warning":  "warning",
		"WARNING":  "warning",
		"error":    "critical",
		"critical": "critical",
		"crit":     "critical",
	}
	for input, expected := range cases {
		got, err := NormalizeStatus(input)
		assert.Nil(t, err)
		assert.Equal(t, expected, got)
	}

	_, err := NormalizeStatus("bogus")
	assert.NotNil(t, err)
}

func TestReadQueryInline(t *testing.T) {
	query, err := ReadQuery("select 'ok', 'fine' from dual;", "")
	assert.Nil(t, err)
	assert.Equal(t, "select 'ok', 'fine' from dual", query)
}

func TestReadQueryPLSQLKeepsTerminator(t *testing.T) {
	query, err := ReadQuery("begin my_proc(:status, :message); end;\n/\n", "")
	assert.Nil(t, err)
	assert.Equal(t, "begin my_proc(:status, :message); end;", query)
}

func TestReadQueryFile(t *testing.T) {
	const filename = "test_query.sql"
	os.WriteFile(filename, []byte("select 'ok', 'fine' from dual\n"), 0644)
	defer os.Remove(filename)

	query, err := ReadQuery("", filename)
	assert.Nil(t, err)
	assert.Equal(t, "select 'ok', 'fine' from dual", query)
}

func TestReadQueryNone(t *testing.T) {
	_, err := ReadQuery("", "")
	assert.NotNil(t, err)
}

func TestReadQueryBoth(t *testing.T) {
	_, err := ReadQuery("select 1 from dual", "somefile.sql")
	assert.NotNil(t, err)
}

func TestReadQueryMissingFile(t *testing.T) {
	_, err := ReadQuery("", "does-not-exist.sql")
	assert.NotNil(t, err)
}

func TestReadQueryEmptyAfterStrip(t *testing.T) {
	_, err := ReadQuery(";", "")
	assert.NotNil(t, err)
}

func TestAggregateQueryOutcomesAllOk(t *testing.T) {
	status, output := AggregateQueryOutcomes([]QueryOutcome{
		{Label: "a", Status: "ok", Message: "fine"},
		{Label: "b", Status: "ok", Message: "fine"},
	})
	assert.Equal(t, "ok", status)
	assert.Equal(t, "0 critical, 0 warning, 2 ok (of 2)", output)
}

func TestAggregateQueryOutcomesWarningWins(t *testing.T) {
	status, output := AggregateQueryOutcomes([]QueryOutcome{
		{Label: "a", Status: "ok", Message: "fine"},
		{Label: "b (u@db)", Status: "warning", Message: "busy"},
	})
	assert.Equal(t, "warning", status)
	assert.Contains(t, output, "0 critical, 1 warning, 1 ok")
	assert.Contains(t, output, "- b (u@db): WARNING busy")
}

func TestAggregateQueryOutcomesCriticalWins(t *testing.T) {
	status, output := AggregateQueryOutcomes([]QueryOutcome{
		{Label: "a", Status: "warning", Message: "busy"},
		{Label: "b", Status: "critical", Message: "down"},
		{Label: "c", Status: "ok", Message: "fine"},
	})
	assert.Equal(t, "critical", status)
	assert.Contains(t, output, "1 critical, 1 warning, 1 ok (of 3)")
	assert.Contains(t, output, "- b: CRITICAL down")
	assert.Contains(t, output, "- a: WARNING busy")
}

func TestAggregateQueryOutcomesUnknownStatusIsCritical(t *testing.T) {
	status, output := AggregateQueryOutcomes([]QueryOutcome{
		{Label: "a", Status: "", Message: "connection failed"},
	})
	assert.Equal(t, "critical", status)
	assert.Contains(t, output, "- a: CRITICAL connection failed")
}

func TestAggregateQueryOutcomesEmpty(t *testing.T) {
	status, output := AggregateQueryOutcomes([]QueryOutcome{})
	assert.Equal(t, "ok", status)
	assert.Equal(t, "0 critical, 0 warning, 0 ok (of 0)", output)
}
