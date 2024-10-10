package data

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
)

type Account struct {
	ID          int64 `json:"id"`
	UserID      int64 `json:"user_id"`
	Balance     int64 `json:"balance"`
	HoldBalance int64 `json:"hold_balance"`
	Version     int   `json:"-"`
}

func ValidateAccount(v *validator.Validator, account *Account) {
	if account.UserID == 0 {
		panic("missing user id for account")
	}

	if account.Version == 0 {
		panic("missing version for account")
	}
}

type AccountModel struct {
	DB *sql.DB
}

func (m AccountModel) Insert(account *Account) error {
	query := `
        INSERT INTO accounts (user_id, balance, hold_balance, version)
		VALUES ($1, $2, $3, $4)
        RETURNING id`

	args := []any{account.UserID, account.Balance, account.HoldBalance, account.Version}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&account.ID)
}

func (m AccountModel) Get(id int64, userId int64) (*Account, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, balance, hold_balance, version
		FROM accounts
        WHERE id = $1
		AND user_id = $2`

	account := Account{
		UserID: userId,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id, userId).Scan(
		&account.ID,
		&account.Balance,
		&account.HoldBalance,
		&account.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &account, nil
}

func (m AccountModel) Update(account *Account) error {
	query := `
        UPDATE accounts
		SET balance = $1, hold_balance = $2, version = version + 1
		WHERE id = $3 AND version = $4 AND user_id = $5
        RETURNING version`

	args := []any{
		account.Balance,
		account.HoldBalance,
		account.ID,
		account.Version,
		account.UserID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&account.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
