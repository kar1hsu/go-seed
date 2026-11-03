package dbx

import (
	"errors"
	"fmt"

	driver "github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
)

// ErrorFields never includes MySQLError.Message, which can contain row values.
func ErrorFields(err error) []logx.LogField {
	fields := []logx.LogField{logx.Field("errorType", fmt.Sprintf("%T", err))}
	var mysqlErr *driver.MySQLError
	if errors.As(err, &mysqlErr) {
		fields = append(fields, logx.Field("mysqlErrno", mysqlErr.Number), logx.Field("sqlState", string(mysqlErr.SQLState[:])))
	}
	return fields
}
