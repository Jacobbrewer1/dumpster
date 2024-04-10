package main

import (
	"cloud.google.com/go/storage"
	"context"
	"flag"
	"fmt"
	"github.com/Jacobbrewer1/dumpster/pkg/logging"
	"google.golang.org/api/option"
	"log/slog"
	"os"
	"time"

	"github.com/Jacobbrewer1/dumpster/pkg/dataaccess"
	"github.com/google/subcommands"
)

type purgeCmd struct {
	// days is the number of days to keep data for. If 0 (or not set), data will not be purged.
	days int

	// gcs is the name of the Google Cloud Storage bucket to use. Setting this will enable GCS.
	gcs string
}

func (p *purgeCmd) Name() string {
	return "purge"
}

func (p *purgeCmd) Synopsis() string {
	return "Purge old dump files"
}

func (p *purgeCmd) Usage() string {
	return `purge:
  Purge old dump files.
`
}

func (p *purgeCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&p.days, "days", 0, "The number of days to keep data for. If 0 (or not set), data will not be purged.")
	f.StringVar(&p.gcs, "gcs", "", "The name of the Google Cloud Storage bucket to use. (Setting this will enable GCS)")
}

func (p *purgeCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// Check if the GCS environment variable is set
	if p.gcs != "" {
		got := os.Getenv(dataaccess.EnvGCSCredentials)
		if got == "" {
			slog.Error("GCS_CREDENTIALS environment variable not set")
			f.Usage()
			return subcommands.ExitUsageError
		}
	}

	// Initialize the GCS client
	var storageClient dataaccess.Storage

	switch {
	case p.gcs != "":
		// Get the service account credentials from the environment variable.
		gcsCredentials := os.Getenv(dataaccess.EnvGCSCredentials)
		if gcsCredentials == "" {
			slog.Error("GCS_CREDENTIALS environment variable not set")
			return subcommands.ExitUsageError
		}

		client, err := storage.NewClient(ctx, option.WithCredentialsJSON([]byte(gcsCredentials)))
		if err != nil {
			slog.Error("error creating GCS client", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}
		cs := client

		_, err = cs.Bucket(p.gcs).Attrs(ctx)
		if err != nil {
			slog.Error("error checking bucket", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}

		storageClient = dataaccess.NewGCS(cs, p.gcs)
		if err != nil {
			slog.Error("error initializing GCS", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}
	default:
		// Locally store the dump
		storageClient = dataaccess.NewLocal()
	}

	// Purge the data
	err := purgeData(ctx, storageClient, p.days)
	if err != nil {
		slog.Error("error purging data", slog.String("error", err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func purgeData(ctx context.Context, r dataaccess.Storage, days int) error {
	if days == 0 {
		slog.Debug("Days to purge is 0, data will not be purged")
		return nil
	}

	// Calculate the date to purge from
	from := time.Now().UTC().AddDate(0, 0, -days)

	// Set the purge date to midnight
	from = time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, from.Location())

	// Check to see if the dumps directory exists
	_, err := os.Stat("dumps")
	if os.IsNotExist(err) {
		slog.Debug("Dumps directory does not exist locally, skipping local purge")
	} else if err != nil {
		return fmt.Errorf("error checking dump directory: %w", err)
	} else {
		// Check local file system for dump files
		files, err := os.ReadDir("dumps")
		if err != nil {
			return fmt.Errorf("error reading dump directory: %w", err)
		}

		// Check if there are any files to purge
		if len(files) == 0 {
			slog.Debug("No files to purge locally")
		} else {
			slog.Debug(fmt.Sprintf("Found %d files to purge locally", len(files)))

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
				if fileDate.After(from) {
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
		}
	}

	// Purge the data
	num, err := r.Purge(ctx, from)
	if err != nil {
		return fmt.Errorf("error purging data from GCS: %w", err)
	}
	slog.Info(fmt.Sprintf("Purged %d files from GCS", num))

	return nil
}
