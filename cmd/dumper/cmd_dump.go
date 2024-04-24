package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"github.com/Jacobbrewer1/dumpster/pkg/dataaccess"
	"github.com/Jacobbrewer1/dumpster/pkg/dumpster"
	"github.com/Jacobbrewer1/dumpster/pkg/logging"
	"github.com/Jacobbrewer1/dumpster/pkg/vault"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/subcommands"
	"github.com/spf13/viper"
	"google.golang.org/api/option"
)

type dumpCmd struct {
	// gcs is the bucket to upload the dump to. Setting this will enable GCS.
	gcs string

	// dbConnStr is the connection string to the database.
	dbConnStr string

	// vaultEnabled is whether to use vault for secrets.
	//
	// This cannot be used in tandem with the dbConnStr flag.
	vaultEnabled bool

	// host is the host of the database. Only used if vault is enabled.
	host string

	// schema is the schema of the database. Only used if vault is enabled.
	schema string

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
	f.BoolVar(&c.vaultEnabled, "vault", false, "Whether to use vault to access the database secrets (Requires VAULT_ADDR, VAULT_APPROLE_ID and VAULT_APPROLE_SECRET_ID environment variables to be set)")
	f.StringVar(&c.host, "host", "", "The host of the database (Only used if vault is enabled)")
	f.StringVar(&c.schema, "schema", "", "The schema of the database (Only used if vault is enabled)")
	f.IntVar(&c.purge, "purge", 0, "The number of days to keep data for. If 0 (or not set), data will not be purged.")
}

func (c *dumpCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	err := logging.Init(appName)
	if err != nil {
		slog.Error("error initializing logging", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	if c.vaultEnabled && c.dbConnStr != "" {
		slog.Error("cannot use vault and db-conn flags together")
		f.Usage()
		return subcommands.ExitUsageError
	}

	// Check if the database connection string is set
	if c.dbConnStr == "" && !c.vaultEnabled {
		slog.Error("database connection string not set")
		f.Usage()
		return subcommands.ExitUsageError
	} else if c.vaultEnabled {
		vip := viper.New()

		err = vip.BindEnv("vault.addr", "VAULT_ADDR")
		if err != nil {
			slog.Error("error binding environment variable", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}

		vc, err := vault.NewClient(vip.GetString("vault.addr"))
		if err != nil {
			slog.Error("error creating vault client", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}

		err = vip.BindEnv("vault.path", "VAULT_DB_PATH")
		if err != nil {
			slog.Error("error binding environment variable", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}

		vs, err := vc.GetSecrets(vip.GetString("vault.path"))
		if err != nil {
			slog.Error("error getting database secrets", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}

		vip.Set("db.host", c.host)
		vip.Set("db.schema", c.schema)

		connStr := dataaccess.GenerateConnectionStr(vip, vs)
		c.dbConnStr = connStr

		slog.Debug("database connection setup from vault")
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

	schemaName, err := d.GetSchemaName()
	if err != nil {
		slog.Error("error getting schema name", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	timestamp := time.Now().UTC().Format(time.RFC3339)
	path := fmt.Sprintf("dumps/%s/%s.sql", schemaName, timestamp)

	var storageClient dataaccess.Storage

	switch {
	case c.gcs != "":
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

		_, err = cs.Bucket(c.gcs).Attrs(ctx)
		if err != nil {
			slog.Error("error checking bucket", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}

		storageClient = dataaccess.NewGCS(cs, c.gcs)
		if err != nil {
			slog.Error("error initializing GCS", slog.String(logging.KeyError, err.Error()))
			return subcommands.ExitFailure
		}
	default:
		// Locally store the dump
		storageClient = dataaccess.NewLocal()
	}

	if err := c.saveDump(ctx, storageClient, fc, path); err != nil {
		slog.Error("error saving dump", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	slog.Info("Dump file created", slog.String("path", path))

	// Purge the data
	if err := purgeData(ctx, storageClient, c.purge); err != nil {
		slog.Error("error purging data", slog.String(logging.KeyError, err.Error()))
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

func (c *dumpCmd) saveDump(ctx context.Context, sc dataaccess.Storage, fileContents string, path string) error {
	// Upload the dump
	err := sc.SaveFile(ctx, path, []byte(fileContents))
	if err != nil {
		return fmt.Errorf("error uploading dump: %w", err)
	}

	return nil
}
