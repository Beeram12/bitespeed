package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// EnsureSchema makes sure that the contacts table and its supporting indexes
// exist in the target database. It is safe to call on every application start.
func EnsureSchema(db *sqlx.DB) error {
	if db == nil {
		return fmt.Errorf("db is nil")
	}

	stmts := []string{
		`
		CREATE TABLE IF NOT EXISTS contacts (
			id BIGSERIAL PRIMARY KEY,
			phone_number TEXT NULL,
			email TEXT NULL,
			linked_id BIGINT NULL REFERENCES contacts(id),
			link_precedence TEXT NOT NULL CHECK (link_precedence IN ('primary','secondary')),
			created_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL,
			deleted_at TIMESTAMPTZ NULL
		);
		`,
		`CREATE INDEX IF NOT EXISTS idx_contacts_email ON contacts(email);`,
		`CREATE INDEX IF NOT EXISTS idx_contacts_phone_number ON contacts(phone_number);`,
		`CREATE INDEX IF NOT EXISTS idx_contacts_linked_id ON contacts(linked_id);`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

