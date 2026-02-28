package contact

import "time"

// LinkPrecedence describes whether a contact is the primary record for a
// cluster of identities or a secondary that points back to a primary.
type LinkPrecedence string

const (
	LinkPrecedencePrimary   LinkPrecedence = "primary"
	LinkPrecedenceSecondary LinkPrecedence = "secondary"
)

type Contact struct {
	ID             int64          `db:"id"`
	PhoneNumber    *string        `db:"phone_number"`
	Email          *string        `db:"email"`
	LinkedID       *int64         `db:"linked_id"`
	LinkPrecedence LinkPrecedence `db:"link_precedence"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
	DeletedAt      *time.Time     `db:"deleted_at"`
}
