package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

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

	// purge is the number of days to keep data for. If 0 (or not set), data will not be purged.
	purge int
}

func (c *dumpCmd) Name() string {
	return "dump"
}

func (c *dumpCmd) Synopsis() string {
	return "Creates a MySQL dump of the database"
}

func (c *dumpCmd) Usage() string {
	return `dump:
  Creates a MySQL dump of the database.
`
}

func (c *dumpCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&c.gcs, "gcs", "", "The GCS bucket to upload the dump to (Requires GCS_CREDENTIALS environment variable to be set)")
	f.StringVar(&c.dbConnStr, "db-conn", "", "The connection string to the database")
	f.IntVar(&c.purge, "purge", 0, "The number of days to keep data for. If 0 (or not set), data will not be purged.")
}

func (c *dumpCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := logging.Init(appName)
	if err != nil {
		slog.Error("error initializing logging", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	// Check if the database connection string is set
	if c.dbConnStr == "" {
		slog.Error("database connection string not set")
		f.Usage()
		return subcommands.ExitUsageError
	}

	// Open database connection
	db, err := sql.Open("mysql", c.dbConnStr)
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
	fc, err := d.Dump()
	if err != nil {
		slog.Error("error creating dump", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	if err := c.uploadDump(ctx, fc); err != nil {
		slog.Error("error uploading dump", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	slog.Info("Dump file created")

	// Purge the data
	if err := c.purgeData(ctx); err != nil {
		slog.Error("error purging data", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (c *dumpCmd) uploadDump(ctx context.Context, fileContents string) error {
	if c.gcs == "" {
		return nil
	}

	if err := dataaccess.ConnectGCS(c.gcs); err != nil {
		return fmt.Errorf("error connecting to GCS: %w", err)
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)

	// Upload the dump
	err := dataaccess.GCS.SaveFile(ctx, fmt.Sprintf("dumps/%s", timestamp), []byte(fileContents))
	if err != nil {
		return fmt.Errorf("error uploading dump: %w", err)
	}

	return nil
}

func (c *dumpCmd) purgeData(ctx context.Context) error {
	if c.purge == 0 {
		slog.Debug("Purge not set, data will not be purged")
		return nil
	}

	// Check local file system for dump files
	files, err := os.ReadDir("dumps")
	if err != nil {
		return fmt.Errorf("error reading dump directory: %w", err)
	}

	// Check if there are any files to purge
	if len(files) == 0 {
		slog.Debug("No files to purge")
		return nil
	}

	localCount := 0

	// Purge the local file system
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		// Parse the file date from the file name
		fileDate, err := time.Parse(time.RFC3339, file.Name())
		if err != nil {
			slog.Warn(fmt.Sprintf("Error parsing file date from file name: %s", file.Name()))
			continue
		}

		// Check if the file date is before the purge date
		if fileDate.After(time.Now().UTC().AddDate(0, 0, -c.purge)) {
			continue
		}

		// Delete the file
		err = os.Remove(fmt.Sprintf("dumps/%s", file.Name()))
		if err != nil {
			return fmt.Errorf("error deleting file: %w", err)
		}

		slog.Info(fmt.Sprintf("Purged file: %s", file.Name()))
		localCount++
	}

	if localCount > 0 {
		slog.Info(fmt.Sprintf("Purged %d files locally", localCount))
	} else {
		slog.Debug("No files to purge locally")
	}

	// Calculate the date to purge from
	from := time.Now().UTC().AddDate(0, 0, -c.purge)

	// Set the purge date to midnight
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	// Purge the data
	num, err := dataaccess.GCS.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data from GCS: %w", err)
	}
	slog.Info(fmt.Sprintf("Purged %d files from GCS", num))

	return nil
}
