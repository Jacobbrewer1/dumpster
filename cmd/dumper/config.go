package main

const (
	appName = `dumpster`
)

type DatabaseConnection struct {
	// ConnStr is the connection string to the database.
	ConnStr string `env:"DUMPSTER_DB_CONN_STR"`
}
