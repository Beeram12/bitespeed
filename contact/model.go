package contact

import "time"

type LinkPrecedence string

// defining enum for 'primary' and 'seconday'
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
