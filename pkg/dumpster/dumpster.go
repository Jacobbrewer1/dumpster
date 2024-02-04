package dumpster

import "database/sql"

type Dumpster struct {
	// db is the database to dump
	db *sql.DB
}

// NewDumpster creates a new dumpster
func NewDumpster(db *sql.DB) *Dumpster {
	return &Dumpster{
		db: db,
	}
}
