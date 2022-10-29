package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/calmitchell617/reserva/internal/validator"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

var AnonymousBank = &Bank{}

type Bank struct {
	Id             int64    `json:"id"`
	Name           string   `json:"name"`
	Email          string   `json:"email"`
	Password       password `json:"-"`
	BalanceInCents int64    `json:"balance_in_cents"`
	Activated      bool     `json:"activated"`
	Frozen         bool     `json:"frozen"`
	Version        int64    `json:"-"`
}

func (u *Bank) IsAnonymous() bool {
	return u == AnonymousBank
}

func ValidateEmail(v *validator.Validator, email string) {
	v.Check(email != "", "email", "must be provided")
	v.Check(validator.Matches(email, validator.EmailRX), "email", "must be a valid email address")
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(utf8.RuneCountInString(password) >= 8, "password", "must be at least 8 characters long")
	v.Check(utf8.RuneCountInString(password) <= 72, "password", "must not be more than 72 characters long")
}

func ValidateBank(v *validator.Validator, bank *Bank) {
	v.Check(bank.Name != "", "name", "must be provided")
	v.Check(utf8.RuneCountInString(bank.Name) <= 500, "name", "must not be more than 500 characters long")

	ValidateEmail(v, bank.Email)

	if bank.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *bank.Password.plaintext)
	}

	if bank.Password.hash == nil {
		panic("missing password hash for bank")
	}
}

type BankModel struct {
	Db *sql.DB
}

func (m BankModel) Insert(bank *Bank) error {
	query := `
        INSERT INTO banks (name, email, password_hash) 
        VALUES ($1, $2, $3)
        RETURNING id, version`

	args := []interface{}{bank.Name, bank.Email, bank.Password.hash}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, args...).Scan(&bank.Id, &bank.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "banks_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}

	return nil
}

func (m BankModel) GetByEmail(email string) (*Bank, error) {
	query := `
        SELECT
					id,
					name,
					email,
					password_hash,
					balance_in_cents,
					activated,
					frozen,
					version
        FROM banks
        WHERE email = $1`

	var bank Bank

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, email).Scan(
		&bank.Id,
		&bank.Name,
		&bank.Email,
		&bank.Password.hash,
		&bank.BalanceInCents,
		&bank.Activated,
		&bank.Frozen,
		&bank.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &bank, nil
}

func (m BankModel) Update(bank *Bank) error {
	query := `
        UPDATE banks 
        SET
					name = $1,
					email = $2,
					password_hash = $3,
					balance_in_cents = $4,
					activated = $5,
					frozen = $6,
					version = version + 1
        WHERE id = $7 AND version = $8
        RETURNING version`

	args := []interface{}{
		bank.Name,
		bank.Email,
		bank.Password.hash,
		bank.BalanceInCents,
		bank.Activated,
		bank.Frozen,
		bank.Id,
		bank.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, args...).Scan(&bank.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "banks_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m BankModel) GetForToken(tokenScope, tokenPlaintext string) (*Bank, error) {
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))

	query := `
        SELECT 
					banks.id,
					banks.name,
					banks.email,
					banks.password_hash,
					banks.balance_in_cents,
					banks.activated,
					banks.frozen,
					banks.version
        FROM banks
        INNER JOIN tokens
        ON banks.id = tokens.bank_id
        WHERE tokens.hash = $1
        AND tokens.scope = $2 
        AND tokens.expiry > $3`

	args := []interface{}{tokenHash[:], tokenScope, time.Now()}

	var bank Bank

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, args...).Scan(
		&bank.Id,
		&bank.Name,
		&bank.Email,
		&bank.Password.hash,
		&bank.BalanceInCents,
		&bank.Activated,
		&bank.Frozen,
		&bank.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &bank, nil
}
