package dumpster

import (
	"bytes"
	"fmt"
	"text/template"
	"time"
)

type ddl struct {
	Database      string
	ServerVersion string
	Tables        []*table
	Triggers      []*trigger
	CompleteTime  string
}

func (d *Dumpster) GetDDL() (string, error) {
	schemaName, err := d.GetSchemaName()
	if err != nil {
		return "", fmt.Errorf("error getting schema name: %w", err)
	}

	data := ddl{
		Database: schemaName,
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

		t.Values = "" // For the DDL we don't need the values

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
