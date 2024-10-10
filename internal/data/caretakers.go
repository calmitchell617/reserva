package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrDuplicateEmail      = errors.New("duplicate email")
	ErrDuplicateExternalID = errors.New("duplicate external ID")
	ErrDuplicateEndpoint   = errors.New("duplicate endpoint")
)

var AnonymousCaretaker = &Caretaker{}

type Caretaker struct {
	ID           int64     `json:"id"`
	ExternalID   string    `json:"external_id"`
	IsAdmin      bool      `json:"is_admin"`
	Endpoint     string    `json:"endpoint"`
	Activated    bool      `json:"activated"`
	Email        string    `json:"email"`
	Password     password  `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	LastModified time.Time `json:"last_modified"`
	Version      int       `json:"-"`
}

func (u *Caretaker) IsAnonymous() bool {
	return u == AnonymousCaretaker
}

type password struct {
	plaintext *string
	hash      []byte
}

func (p *password) Set(plaintextPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), 12)
	if err != nil {
		return err
	}

	p.plaintext = &plaintextPassword
	p.hash = hash

	return nil
}

func (p *password) Matches(plaintextPassword string) (bool, error) {
	err := bcrypt.CompareHashAndPassword(p.hash, []byte(plaintextPassword))
	if err != nil {
		switch {
		case errors.Is(err, bcrypt.ErrMismatchedHashAndPassword):
			return false, nil
		default:
			return false, err
		}
	}

	return true, nil
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(len(password) >= 8, "password", "must be at least 8 bytes long")
	v.Check(len(password) <= 72, "password", "must not be more than 72 bytes long")
}

func ValidateCaretaker(v *validator.Validator, caretaker *Caretaker) {
	v.Check(caretaker.ExternalID != "", "external_id", "must be provided")
	v.Check(caretaker.Endpoint != "", "endpoint", "must be provided")

	ValidateEmail(v, caretaker.Email)

	if caretaker.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *caretaker.Password.plaintext)
	}

	if caretaker.Password.hash == nil {
		panic("missing password hash for caretaker")
	}

	if caretaker.CreatedAt.IsZero() {
		panic("missing created at timestamp for caretaker")
	}

	if caretaker.LastModified.IsZero() {
		panic("missing last modified timestamp for caretaker")
	}

	if caretaker.Version == 0 {
		panic("missing version for caretaker")
	}
}

type CaretakerModel struct {
	DB *sql.DB
}

func (m CaretakerModel) Insert(caretaker *Caretaker) error {
	query := `INSERT INTO caretakers(external_id, is_admin, endpoint, activated, email, password_hash, created_at, last_modified, version)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING id`

	args := []any{caretaker.ExternalID, caretaker.IsAdmin, caretaker.Endpoint, caretaker.Activated, caretaker.Email, caretaker.Password.hash, caretaker.CreatedAt, caretaker.LastModified, caretaker.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&caretaker.ID)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "caretakers_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "caretakers_external_id_key"`:
			return ErrDuplicateExternalID
		case err.Error() == `pq: duplicate key value violates unique constraint "caretakers_endpoint_key"`:
			return ErrDuplicateEndpoint
		default:
			return err
		}
	}

	return nil
}

func (m CaretakerModel) GetByEmail(email string) (*Caretaker, error) {
	query := `SELECT id, external_id, is_admin, endpoint, activated, email, password_hash, created_at, last_modified, version FROM caretakers WHERE email = $1`

	var caretaker Caretaker

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, email).Scan(
		&caretaker.ID,
		&caretaker.ExternalID,
		&caretaker.IsAdmin,
		&caretaker.Endpoint,
		&caretaker.Activated,
		&caretaker.Email,
		&caretaker.Password.hash,
		&caretaker.CreatedAt,
		&caretaker.LastModified,
		&caretaker.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &caretaker, nil
}

func (m CaretakerModel) Update(caretaker *Caretaker) error {
	query := `
        UPDATE caretakers 
		SET external_id = $1, is_admin = $2, endpoint = $3, activated = $4, email = $5, password_hash = $6, last_modified = $7, version = version + 1
		WHERE id = $8 AND version = $9
        RETURNING version`

	args := []any{
		caretaker.ExternalID,
		caretaker.IsAdmin,
		caretaker.Endpoint,
		caretaker.Activated,
		caretaker.Email,
		caretaker.Password.hash,
		caretaker.LastModified,
		caretaker.ID,
		caretaker.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&caretaker.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "caretakers_email_key"`:
			return ErrDuplicateEmail
		case err.Error() == `pq: duplicate key value violates unique constraint "caretakers_external_id_key"`:
			return ErrDuplicateExternalID
		case err.Error() == `pq: duplicate key value violates unique constraint "caretakers_endpoint_key"`:
			return ErrDuplicateEndpoint
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m CaretakerModel) GetForToken(tokenScope, tokenPlaintext string) (*Caretaker, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT caretakers.id, caretakers.external_id, caretakers.is_admin, caretakers.endpoint, caretakers.activated, caretakers.email, caretakers.password_hash, caretakers.created_at, caretakers.last_modified, caretakers.version
        FROM caretakers
        INNER JOIN tokens
        ON caretakers.id = tokens.caretaker_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []any{tokenHash[:], tokenScope, time.Now()}

	var caretaker Caretaker

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&caretaker.ID,
		&caretaker.ExternalID,
		&caretaker.IsAdmin,
		&caretaker.Endpoint,
		&caretaker.Activated,
		&caretaker.Email,
		&caretaker.Password.hash,
		&caretaker.CreatedAt,
		&caretaker.LastModified,
		&caretaker.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &caretaker, nil
}
