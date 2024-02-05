package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/Jacobbrewer1/dumpster/pkg/dataaccess"
	"github.com/Jacobbrewer1/dumpster/pkg/dumpster"
	"github.com/Jacobbrewer1/dumpster/pkg/logging"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/subcommands"
)

type dumpCmd struct {
	// gcs is the bucket to upload the dump to. Setting this will enable GCS.
	gcs string

	// dbConnStr is the connection string to the database.
	dbConnStr string
}

func (m *dumpCmd) Name() string {
	return "dump"
}

func (m *dumpCmd) Synopsis() string {
	return "Creates a MySQL dump of the database"
}

func (m *dumpCmd) Usage() string {
	return `dump:
Creates a MySQL dump of the database.
`
}

func (m *dumpCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&m.gcs, "gcs", "", "The GCS bucket to upload the dump to (Requires GCS_CREDENTIALS environment variable to be set)")
	f.StringVar(&m.dbConnStr, "db-conn", "", "The connection string to the database")
}

func (m *dumpCmd) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := logging.Init(appName)
	if err != nil {
		slog.Error("error initializing logging", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	// Check if the database connection string is set
	if m.dbConnStr == "" {
		slog.Error("database connection string not set")
		return subcommands.ExitUsageError
	}

	// Open database connection
	db, err := sql.Open("mysql", m.dbConnStr)
	if err != nil {
		slog.Error("error opening database", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	// Close the database connection
	defer func(db *sql.DB) {
		if err := db.Close(); err != nil {
			slog.Warn("error closing database: %v", err)
		}
	}(db)

	// Create a new dumpster
	d := dumpster.NewDumpster(db)

	// Create the dump
	f, err := d.Dump()
	if err != nil {
		slog.Error("error creating dump", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	if err := m.uploadDump(f); err != nil {
		slog.Error("error uploading dump", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (m *dumpCmd) uploadDump(pathToFile string) error {
	if m.gcs == "" {
		return nil
	}

	if err := dataaccess.ConnectGCS(m.gcs); err != nil {
		return fmt.Errorf("error connecting to GCS: %w", err)
	}

	// Get the file from the file system
	file, err := os.ReadFile(pathToFile)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	// Only keep the file name
	_, fileName := filepath.Split(pathToFile)

	// Upload the dump
	err = dataaccess.GCS.SaveFile(context.Background(), fmt.Sprintf("dumps/%s", fileName), file)
	if err != nil {
		return fmt.Errorf("error uploading dump: %w", err)
	}

	return nil
}
