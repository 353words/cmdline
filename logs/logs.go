package logs

import (
	"time"

	_ "github.com/duckdb/duckdb-go/v2"
	"github.com/jmoiron/sqlx"
)

// Record is a log record.
type Record struct {
	Origin     string
	Time       time.Time
	Method     string
	Path       string
	StatusCode int `db:"status_code"`
	Size       int
}

// Filter is a query filter.
type Filter struct {
	Path string
	// TODO: Other filter fields
}

func Query(conn *sqlx.DB, filter Filter) ([]Record, error) {
	var records []Record

	const querySQL = `SELECT * FROM logs WHERE path LIKE ?`
	pathQuery := "%" + filter.Path + "%"
	if err := conn.Select(&records, querySQL, pathQuery); err != nil {
		return nil, err
	}

	return records, nil
}
