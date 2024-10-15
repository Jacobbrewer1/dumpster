package dumpster

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"strings"
	"text/template"
	"time"

	"github.com/Jacobbrewer1/dumpster/pkg/logging"
)

type table struct {
	Name   string
	SQL    string
	Values string
}

type trigger struct {
	Name string
	SQL  string
}

type dump struct {
	Database      string
	ServerVersion string
	Tables        []*table
	Triggers      []*trigger
	CompleteTime  string
}

// DumpFile creates a new dump of the database
func (d *Dumpster) DumpFile() (string, error) {
	timestamp := time.Now().Format(time.RFC3339)

	data, err := d.Dump()
	if err != nil {
		return "", fmt.Errorf("error creating dump: %w", err)
	}

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
			slog.Warn("Error closing file", slog.String(logging.KeyError, err.Error()))
		}
	}(f)

	// Write the dump to the file
	if _, err := f.WriteString(data); err != nil {
		return "", fmt.Errorf("error writing to file: %w", err)
	}

	return p, nil
}

// Dump creates a new dump of the database and returns the content.
func (d *Dumpster) Dump() (string, error) {
	schemaName, err := d.GetSchemaName()
	if err != nil {
		return "", fmt.Errorf("error getting schema name: %w", err)
	}

	data := dump{
		Database: schemaName,
		Tables:   make([]*table, 0),
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

	// Get triggers
	triggers, err := d.getTriggers()
	if err != nil {
		return "", fmt.Errorf("error getting triggers: %w", err)
	}

	// Get sql for each trigger
	for _, tn := range triggers {
		t, err := d.createTrigger(tn)
		if err != nil {
			return "", fmt.Errorf("error creating trigger: %w", err)
		}

		data.Triggers = append(data.Triggers, t)
	}

	// Set complete time
	data.CompleteTime = time.Now().Format(time.RFC3339)

	t, err := template.New("mysqldump").Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	b := new(bytes.Buffer)
	if err = t.Execute(b, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return b.String(), nil
}

func (d *Dumpster) getTriggers() ([]string, error) {
	sqlStmt := "SHOW TRIGGERS"

	// Prepare statement for reading data
	stmt, err := d.db.Prepare(sqlStmt)
	if err != nil {
		return nil, fmt.Errorf("error preparing statement: %w", err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Warn("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}
	}(stmt)

	// Execute statement
	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Warn("Error closing rows", slog.String(logging.KeyError, err.Error()))
		}
	}(rows)

	// Read data
	triggers := make([]string, 0)
	for rows.Next() {
		t := new(sql.NullString)
		event := new(sql.NullString)
		sqlTable := new(sql.NullString)
		statement := new(sql.NullString)
		timing := new(sql.NullString)
		created := new(sql.NullString)
		sqlMode := new(sql.NullString)
		definer := new(sql.NullString)
		characterSetClient := new(sql.NullString)
		collationConnection := new(sql.NullString)
		databaseCollation := new(sql.NullString)

		if err := rows.Scan(t, event, sqlTable, statement, timing, created, sqlMode, definer,
			characterSetClient, collationConnection, databaseCollation); err != nil {
			return nil, fmt.Errorf("error scanning: %w", err)
		}

		if t.Valid {
			triggers = append(triggers, t.String)
		} else {
			slog.Warn("trigger is not valid", slog.String("trigger", t.String))
		}
	}

	return triggers, nil
}

func (d *Dumpster) createTrigger(name string) (t *trigger, err error) {
	t = &trigger{
		Name: name,
	}

	if t.SQL, err = d.createTriggerSQL(name); err != nil {
		return nil, err
	}

	return t, nil
}

func (d *Dumpster) createTriggerSQL(name string) (string, error) {
	sqlStmt := "SHOW CREATE TRIGGER " + name

	// Prepare statement for reading data
	stmt, err := d.db.Prepare(sqlStmt)
	if err != nil {
		return "", fmt.Errorf("error preparing statement: %w", err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Warn("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}
	}(stmt)

	// Execute statement
	triggerName := new(sql.NullString)
	sqlMode := new(sql.NullString)
	originalStatement := new(sql.NullString)
	characterSetClient := new(sql.NullString)
	collationConnection := new(sql.NullString)
	databaseCollation := new(sql.NullString)
	createdAt := new(sql.NullString)

	if err := stmt.QueryRow().Scan(triggerName, sqlMode, originalStatement, characterSetClient,
		collationConnection, databaseCollation, createdAt); err != nil {
		return "", fmt.Errorf("error executing statement: %w", err)
	}

	if originalStatement.Valid {
		return originalStatement.String, nil
	} else {
		return "", fmt.Errorf("error getting trigger SQL: %w", err)
	}
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
			slog.Warn("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}
	}(stmt)

	// Execute statement
	rows, err := stmt.Query()
	if err != nil {
		return nil, fmt.Errorf("error executing statement: %w", err)
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Warn("Error closing rows", slog.String(logging.KeyError, err.Error()))
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
			slog.Warn("table is not valid", slog.String("table", t.String))
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
			slog.Warn("Error closing statement", slog.String(logging.KeyError, err.Error()))
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
			slog.Warn("Error closing statement", slog.String(logging.KeyError, err.Error()))
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
			slog.Warn("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}
	}(stmt)

	// Execute statement
	rows, err := stmt.Query()
	if err != nil {
		return "", fmt.Errorf("error executing statement: %w", err)
	}

	defer func(rows *sql.Rows) {
		if err := rows.Close(); err != nil {
			slog.Warn("Error closing rows", slog.String(logging.KeyError, err.Error()))
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

func (d *Dumpster) GetSchemaName() (string, error) {
	sqlStmt := "SELECT DATABASE()"

	// Prepare statement for reading data
	stmt, err := d.db.Prepare(sqlStmt)
	if err != nil {
		return "", fmt.Errorf("error preparing statement: %w", err)
	}

	defer func(stmt *sql.Stmt) {
		if err := stmt.Close(); err != nil {
			slog.Warn("Error closing statement", slog.String(logging.KeyError, err.Error()))
		}
	}(stmt)

	// Execute statement
	var schema sql.NullString
	if err := stmt.QueryRow().Scan(&schema); err != nil {
		return "", fmt.Errorf("error executing statement: %w", err)
	}

	if !schema.Valid {
		return "", errors.New("returned schema is not valid")
	}

	return schema.String, nil
}
