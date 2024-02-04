package dumpster

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"text/template"
	"time"
)

type table struct {
	Name   string
	SQL    string
	Values string
}

type dump struct {
	ServerVersion string
	Tables        []*table
	CompleteTime  string
}

// Dump creates a new dump of the database
func (d *Dumpster) Dump() (string, error) {
	timestamp := time.Now().Format(time.RFC3339)

	timestamp = strings.Replace(timestamp, ":", "-", -1)

	// Get the PWD
	pwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting the current working directory: %w", err)
	}

	pwd = pwd + "/dumps"

	// Create the dump directory
	p := path.Join(pwd, timestamp+".sql")

	// Ensure that the full path exists, if not create it
	if err := os.MkdirAll(pwd, os.ModePerm); err != nil {
		return "", fmt.Errorf("error creating dump directory: %w", err)
	}

	// Create .sql file
	f, err := os.Create(p)
	if err != nil {
		return "", fmt.Errorf("error creating file: %w", err)
	}

	defer func(f *os.File) {
		if err := f.Close(); err != nil {
			slog.Warn("error closing file: %v", err)
		}
	}(f)

	data := dump{
		Tables: make([]*table, 0),
	}

	// Get server version
	if data.ServerVersion, err = d.getServerVersion(); err != nil {
		return "", fmt.Errorf("error getting server version: %w", err)
	}

	// Get tables
	tables, err := d.getTables()
	if err != nil {
		return "", fmt.Errorf("error getting tables: %w", err)
	}

	// Get sql for each table
	for _, tn := range tables {
		t, err := d.createTable(tn)
		if err != nil {
			return "", fmt.Errorf("error creating table: %w", err)
		}

		data.Tables = append(data.Tables, t)
	}

	// Set complete time
	data.CompleteTime = time.Now().Format(time.RFC3339)

	// Write dump to file
	t, err := template.New("mysqldump").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	if err = t.Execute(f, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return p, nil
}

func (d *Dumpster) getTables() ([]string, error) {
	sqlStmt := "SHOW TABLES"

	// Prepare statement for reading data
	stmt, err := d.db.Prepare(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Warn("error closing statement: %v", err)
		}
	}(stmt)

	// Execute statement
	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Warn("error closing rows: %v", err)
		}
	}(rows)

	// Read data
	tables := make([]string, 0)
	for rows.Next() {
		var t sql.NullString
		if err := rows.Scan(&t); err != nil {
			return nil, fmt.Errorf("error scanning: %w", err)
		}

		if t.Valid {
			tables = append(tables, t.String)
		} else {
			slog.Warn("table is not valid: %v", t)
		}
	}

	return tables, nil
}

func (d *Dumpster) getServerVersion() (string, error) {
	sqlStmt := "SELECT version()"

	// Prepare statement for reading data
	stmt, err := d.db.Prepare(sqlStmt)
	if err != nil {
		return "", fmt.Errorf("error preparing statement: %w", err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Warn("error closing statement: %v", err)
		}
	}(stmt)

	version := ""

	// Execute statement
	if err := stmt.QueryRow().Scan(&version); err != nil {
		return "", fmt.Errorf("error executing statement: %w", err)
	}

	if version == "" {
		return "", errors.New("returned version is empty")
	}

	return version, nil
}

func (d *Dumpster) createTable(name string) (t *table, err error) {
	t = &table{
		Name: name,
	}

	if t.SQL, err = d.createTableSQL(name); err != nil {
		return nil, err
	}

	if t.Values, err = d.createTableValues(name); err != nil {
		return nil, err
	}

	return t, nil
}

func (d *Dumpster) createTableSQL(name string) (string, error) {
	sqlStmt := "SHOW CREATE TABLE " + name

	// Prepare statement for reading data
	stmt, err := d.db.Prepare(sqlStmt)
	if err != nil {
		return "", fmt.Errorf("error preparing statement: %w", err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Warn("error closing statement: %v", err)
		}
	}(stmt)

	// Execute statement
	var tableReturn sql.NullString
	var tableSql sql.NullString
	if err := stmt.QueryRow().Scan(&tableReturn, &tableSql); err != nil {
		return "", fmt.Errorf("error executing statement: %w", err)
	}

	if tableReturn.String != name {
		return "", errors.New("returned table is not the same as requested table")
	}

	if !tableSql.Valid {
		return "", errors.New("returned table SQL is not valid")
	}

	return tableSql.String, nil
}

func (d *Dumpster) createTableValues(name string) (string, error) {
	sqlStmt := "SELECT * FROM " + name

	// Prepare statement for reading data
	stmt, err := d.db.Prepare(sqlStmt)
	if err != nil {
		return "", fmt.Errorf("error preparing statement: %w", err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Warn("error closing statement: %v", err)
		}
	}(stmt)

	// Execute statement
	rows, err := stmt.Query()
	if err != nil {
		return "", fmt.Errorf("error executing statement: %w", err)
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Warn("error closing rows: %v", err)
		}
	}(rows)

	// Get columns
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("error getting columns: %w", err)
	} else if len(columns) == 0 {
		return "", errors.New("no columns found")
	}

	// Read data
	dataText := make([]string, 0)
	for rows.Next() {
		data := make([]*sql.NullString, len(columns))
		pointers := make([]any, len(columns))
		for i := range data {
			pointers[i] = &data[i]
		}

		// Read data
		if err := rows.Scan(pointers...); err != nil {
			return "", err
		}

		dataStrings := make([]string, len(columns))

		for key, value := range data {
			if value != nil && value.Valid {
				dataStrings[key] = "'" + value.String + "'"
			} else {
				dataStrings[key] = "null"
			}
		}

		dataText = append(dataText, "("+strings.Join(dataStrings, ",")+")")
	}

	return strings.Join(dataText, ","), nil
}
