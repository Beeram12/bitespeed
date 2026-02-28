package contact

import (
	"context"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/jmoiron/sqlx/reflectx"
)

type Repository interface {
	WithTx(tx *sqlx.Tx) Repository
	FindByEmailOrPhone(ctx context.Context, email *string, phone *string) ([]Contact, error)
	FindByIDs(ctx context.Context, ids []int64) ([]Contact, error)
	Create(ctx context.Context, c *Contact) error
	Update(ctx context.Context, c *Contact) error
}

type repo struct {
	db *sqlx.DB
	tx *sqlx.Tx
}

func NewRepository(db *sqlx.DB) Repository {
	db.Mapper = reflectx.NewMapperFunc("db", sqlx.NameMapper)
	return &repo{db: db}
}

func (r *repo) use() sqlx.ExtContext {
	if r.tx != nil {
		return r.tx
	}
	return r.db
}

// WithTx returns a Repository that uses the provided SQL transaction.
func (r *repo) WithTx(tx *sqlx.Tx) Repository {
	return &repo{db: r.db, tx: tx}
}

// FindByEmailOrPhone finds list of contacts using email and phone number
func (r *repo) FindByEmailOrPhone(ctx context.Context, email *string, phone *string) ([]Contact, error) {
	args := []interface{}{}
	clauses := []string{}

	if email != nil {
		clauses = append(clauses, "email = ?")
		args = append(args, *email)
	}
	if phone != nil {
		clauses = append(clauses, "phone_number = ?")
		args = append(args, *phone)
	}

	if len(clauses) == 0 {
		return nil, errors.New("at least one of email or phone must be non-nil")
	}

	query := `
		SELECT id, phone_number, email, linked_id, link_precedence, created_at, updated_at, deleted_at
		FROM contacts
		WHERE ` + clauses[0]

	if len(clauses) == 2 {
		query += " OR " + clauses[1]
	}

	query = r.db.Rebind(query)

	var contacts []Contact
	if err := sqlx.SelectContext(ctx, r.use(), &contacts, query, args...); err != nil {
		return nil, err
	}
	return contacts, nil
}

// FindByIds finds contacts using their ids
func (r *repo) FindByIDs(ctx context.Context, ids []int64) ([]Contact, error) {
	if len(ids) == 0 {
		return []Contact{}, nil
	}

	query, args, err := sqlx.In(`
		SELECT id, phone_number, email, linked_id, link_precedence, created_at, updated_at, deleted_at
		FROM contacts
		WHERE id IN (?)
	`, ids)
	if err != nil {
		return nil, err
	}
	query = r.db.Rebind(query)

	var contacts []Contact
	if err := sqlx.SelectContext(ctx, r.use(), &contacts, query, args...); err != nil {
		return nil, err
	}
	return contacts, nil
}

func (r *repo) Create(ctx context.Context, c *Contact) error {
	query := `
		INSERT INTO contacts (phone_number, email, linked_id, link_precedence, created_at, updated_at, deleted_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		RETURNING id
	`
	query = r.db.Rebind(query)

	args := []interface{}{
		c.PhoneNumber,
		c.Email,
		c.LinkedID,
		c.LinkPrecedence,
		c.CreatedAt,
		c.UpdatedAt,
		c.DeletedAt,
	}

	rows, err := r.use().QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		if err := rows.Scan(&c.ID); err != nil {
			return err
		}
		return nil
	}

	return errors.New("no id returned from insert")
}

func (r *repo) Update(ctx context.Context, c *Contact) error {
	if c.ID == 0 {
		return errors.New("id is required for update")
	}

	query := `
		UPDATE contacts
		SET phone_number = ?,
		    email = ?,
		    linked_id = ?,
		    link_precedence = ?,
		    updated_at = ?,
		    deleted_at = ?
		WHERE id = ?
	`

	query = r.db.Rebind(query)

	args := []interface{}{
		c.PhoneNumber,
		c.Email,
		c.LinkedID,
		c.LinkPrecedence,
		c.UpdatedAt,
		c.DeletedAt,
		c.ID,
	}

	_, err := r.use().ExecContext(ctx, query, args...)
	return err
}
