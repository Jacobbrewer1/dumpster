package dumpster

import (
	"github.com/jmoiron/sqlx"
)

type Dumpster struct {
	// db is the database to dump
	db *sqlx.DB
}

// NewDumpster creates a new dumpster
func NewDumpster(db *sqlx.DB) *Dumpster {
	return &Dumpster{
		db: db,
	}
}
