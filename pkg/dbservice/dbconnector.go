package dbservice

import (
	"database/sql"
)

var DB DBConnector

type DBConnector interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	QueryRow(query string, args ...interface{}) SqlRow
	Query(query string, args ...interface{}) (*sql.Rows, error)
	Close() error
}

type DBConnectorImpl struct {
	DB *sql.DB
}

func (d *DBConnectorImpl) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(query, args...)
}
func (d *DBConnectorImpl) QueryRow(query string, args ...interface{}) SqlRow {
	return &SqlRowImpl{d.DB.QueryRow(query, args...)}
}
func (d *DBConnectorImpl) Close() error {
	return d.DB.Close()
}
func (d *DBConnectorImpl) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.Query(query, args...)
}

type SqlRow interface {
	Err() error
	Scan(dest ...any) error
}

type SqlRowImpl struct {
	row *sql.Row
}

func (r *SqlRowImpl) Err() error {
	return r.row.Err()
}
func (r *SqlRowImpl) Scan(dest ...any) error {
	return r.row.Scan(dest...)
}
