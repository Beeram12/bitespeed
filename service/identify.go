package service

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/Beeram12/bitespeed-identify/contact"
	"github.com/jmoiron/sqlx"
)

type Store interface {
	DB() *sqlx.DB
}

type ContactRepository interface {
	WithTx(tx *sqlx.Tx) contact.Repository
	FindByEmailOrPhone(ctx context.Context, email *string, phone *string) ([]contact.Contact, error)
	FindByIDs(ctx context.Context, ids []int64) ([]contact.Contact, error)
	Create(ctx context.Context, c *contact.Contact) error
	Update(ctx context.Context, c *contact.Contact) error
}

type IdentifyService struct {
	db   *sqlx.DB
	repo contact.Repository
}

func NewIdentifyService(db *sqlx.DB, repo contact.Repository) *IdentifyService {
	return &IdentifyService{
		db:   db,
		repo: repo,
	}
}

type IdentifyRequest struct {
	Email       *string
	PhoneNumber *string
}

type IdentifyResponse struct {
	PrimaryContactID    int64
	Emails              []string
	PhoneNumbers        []string
	SecondaryContactIDs []int64
}

func (s *IdentifyService) Identify(ctx context.Context, req IdentifyRequest) (*IdentifyResponse, error) {
	if req.Email == nil && req.PhoneNumber == nil {
		return nil, errors.New("either email or phoneNumber is required")
	}

	tx, err := s.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	repo := s.repo.WithTx(tx)

	existing, err := repo.FindByEmailOrPhone(ctx, req.Email, req.PhoneNumber)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()

	// case 1: No exisiting contacts, create new primary
	if len(existing) == 0 {
		newContact := &contact.Contact{
			PhoneNumber:    req.PhoneNumber,
			Email:          req.Email,
			LinkedID:       nil,
			LinkPrecedence: contact.LinkPrecedencePrimary,
			CreatedAt:      now,
			UpdatedAt:      now,
			DeletedAt:      nil,
		}
		if err := repo.Create(ctx, newContact); err != nil {
			return nil, err
		}
		if err := tx.Commit(); err != nil {
			return nil, err
		}
		tx = nil

		emails := []string{}
		phones := []string{}
		if newContact.Email != nil {
			emails = append(emails, *newContact.Email)
		}
		if newContact.PhoneNumber != nil {
			phones = append(phones, *newContact.PhoneNumber)
		}

		return &IdentifyResponse{
			PrimaryContactID:    newContact.ID,
			Emails:              emails,
			PhoneNumbers:        phones,
			SecondaryContactIDs: []int64{},
		}, nil
	}

	// case 2: There are exisiting contacts
	all := make([]contact.Contact, len(existing))
	copy(all, existing)

	// find old primary by createdAt or lowestId if there is a tie
	var primary contact.Contact
	var primaries []contact.Contact
	for _, c := range all {
		if c.LinkPrecedence == contact.LinkPrecedencePrimary {
			primaries = append(primaries, c)
		}
	}

	if len(primaries) == 0 {
		// if no primary exisits pick oldest by createdAt as primary
		primaries = append(primaries, all...)
	}

	sort.Slice(primaries, func(i, j int) bool {
		if primaries[i].CreatedAt.Equal(primaries[j].CreatedAt) {
			return primaries[i].ID < primaries[j].ID
		}
		return primaries[i].CreatedAt.Before(primaries[j].CreatedAt)
	})
	primary = primaries[0]

	// change all others to secondary pointing to primary.
	for i := range all {
		if all[i].ID == primary.ID {
			continue
		}
		changed := false
		if all[i].LinkPrecedence != contact.LinkPrecedenceSecondary {
			all[i].LinkPrecedence = contact.LinkPrecedenceSecondary
			changed = true
		}
		if all[i].LinkedID == nil || *all[i].LinkedID != primary.ID {
			all[i].LinkedID = &primary.ID
			changed = true
		}
		if changed {
			all[i].UpdatedAt = now
			if err := repo.Update(ctx, &all[i]); err != nil {
				return nil, err
			}
		}
	}

	// decide if we need to created a new secondary for new info
	emailSet := map[string]struct{}{}
	phoneSet := map[string]struct{}{}

	for _, c := range all {
		if c.Email != nil {
			emailSet[*c.Email] = struct{}{}
		}
		if c.PhoneNumber != nil {
			phoneSet[*c.PhoneNumber] = struct{}{}
		}
	}

	needNew := false
	if req.Email != nil {
		if _, ok := emailSet[*req.Email]; !ok {
			needNew = true
		}
	}
	if req.PhoneNumber != nil {
		if _, ok := phoneSet[*req.PhoneNumber]; !ok {
			needNew = true
		}
	}

	if needNew {
		newSecondary := &contact.Contact{
			PhoneNumber:    req.PhoneNumber,
			Email:          req.Email,
			LinkedID:       &primary.ID,
			LinkPrecedence: contact.LinkPrecedenceSecondary,
			CreatedAt:      now,
			UpdatedAt:      now,
			DeletedAt:      nil,
		}
		if err := repo.Create(ctx, newSecondary); err != nil {
			return nil, err
		}
		all = append(all, *newSecondary)

		// Update sets.
		if newSecondary.Email != nil {
			emailSet[*newSecondary.Email] = struct{}{}
		}
		if newSecondary.PhoneNumber != nil {
			phoneSet[*newSecondary.PhoneNumber] = struct{}{}
		}
	}

	// build response
	emails := []string{}
	phones := []string{}

	// Primary first if present.
	if primary.Email != nil {
		emails = append(emails, *primary.Email)
	}
	if primary.PhoneNumber != nil {
		phones = append(phones, *primary.PhoneNumber)
	}

	// Add other uniques.
	for e := range emailSet {
		if primary.Email != nil && e == *primary.Email {
			continue
		}
		emails = append(emails, e)
	}
	for p := range phoneSet {
		if primary.PhoneNumber != nil && p == *primary.PhoneNumber {
			continue
		}
		phones = append(phones, p)
	}

	secondaryIDs := []int64{}
	for _, c := range all {
		if c.ID == primary.ID {
			continue
		}
		if c.LinkPrecedence == contact.LinkPrecedenceSecondary {
			secondaryIDs = append(secondaryIDs, c.ID)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	tx = nil

	return &IdentifyResponse{
		PrimaryContactID:    primary.ID,
		Emails:              emails,
		PhoneNumbers:        phones,
		SecondaryContactIDs: secondaryIDs,
	}, nil

}
