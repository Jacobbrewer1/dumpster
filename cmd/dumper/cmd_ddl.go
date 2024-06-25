package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Jacobbrewer1/dumpster/pkg/dumpster"
	"github.com/Jacobbrewer1/dumpster/pkg/logging"
	"github.com/caarlos0/env/v11"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/subcommands"
	"github.com/jmoiron/sqlx"
)

type ddlCmd struct{}

func (c *ddlCmd) Name() string {
	return "ddl"
}

func (c *ddlCmd) Synopsis() string {
	return "Creates a MySQL DDL of the database"
}

func (c *ddlCmd) Usage() string {
	return `ddl:
  Creates a MySQL DDL of the database.
`
}

func (c *ddlCmd) SetFlags(f *flag.FlagSet) {}

func (c *ddlCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	dbConnEnv := new(DatabaseConnection)
	if err := env.Parse(dbConnEnv); err != nil {
		slog.Error("error parsing environment variables", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	db, err := sqlx.Connect("mysql", dbConnEnv.ConnStr)
	if err != nil {
		slog.Error("error connecting to database", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	defer func() {
		if err := db.Close(); err != nil {
			slog.Warn("error closing database: %v", slog.String("error", err.Error()))
		}
	}()

	d := dumpster.NewDumpster(db)

	ddlStr, err := d.GetDDL()
	if err != nil {
		slog.Error("error getting DDL", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	schemaName, err := d.GetSchemaName()
	if err != nil {
		slog.Error("error getting schema name", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	path := fmt.Sprintf("ddl/%s.sql", schemaName)

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		slog.Error("error creating directory", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	file, err := os.Create(path)
	if err != nil {
		slog.Error("error creating file", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("error closing file: %v", slog.String("error", err.Error()))
		}
	}()

	_, err = file.WriteString(ddlStr)
	if err != nil {
		slog.Error("error writing DDL to file", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
