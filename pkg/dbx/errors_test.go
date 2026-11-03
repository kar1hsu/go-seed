package dbx

import (
	"fmt"
	driver "github.com/go-sql-driver/mysql"
	"strings"
	"testing"
)

func TestErrorFieldsExcludeRowValues(t *testing.T) {
	err := fmt.Errorf("wrap: %w", &driver.MySQLError{Number: 1270, SQLState: [5]byte{'H', 'Y', '0', '0', '0'}, Message: "secret-value"})
	fields := ErrorFields(err)
	values := map[string]any{}
	for _, f := range fields {
		values[f.Key] = f.Value
	}
	if len(values) != 3 || values["mysqlErrno"] != uint16(1270) || values["sqlState"] != "HY000" || values["errorType"] == nil || strings.Contains(fmt.Sprint(fields), "secret-value") {
		t.Fatalf("invalid error fields: %v", fields)
	}
}
