package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
	"github.com/calmitchell617/reserva/pkg"
	"github.com/go-redis/redis/v8"
)

// import (
// 	"context"
// 	"database/sql"
// 	"errors"
// 	"fmt"
// 	"time"

// 	"github.com/calmitchell617/reserva/internal/validator"
// )

type Account struct {
	Id              int64  `json:"id"`
	ControllingBank string `json:"controlling_bank"`
	Metadata        string `json:"metadata"`
	BalanceInCents  int64  `json:"balance_in_cents"`
	Frozen          bool   `json:"frozen"`
	Version         int64  `json:"version"`
}

func ValidateAccount(v *validator.Validator, account *Account) {
	v.Check(account.ControllingBank != "", "controlling_bank", "must be provided")
	v.Check(account.Metadata != "", "metadata", "must be provided")
	v.Check(pkg.IsJSON(account.Metadata), "metadata", "must be valid JSON")
}

func ValidateAccountMetadata(v *validator.Validator, metadata string) {
	v.Check(pkg.IsJSON(metadata), "controlling_bank", "must be provided")
}

type AccountModel struct {
	Db    *sql.DB
	Cache *redis.Client
}

func (m AccountModel) Insert(account *Account) (int64, error) {
	query := `
        INSERT INTO accounts (controlling_bank, metadata)
        VALUES (?, ?)`

	args := []interface{}{account.ControllingBank, account.Metadata}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := m.Db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("unable to run bank insertion query, err: %v", err)
	}

	accountId, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("unable to get bankId for last account insertion, err: %v", err)
	}

	err = m.Cache.Set(ctx, fmt.Sprintf("accounts/%v", accountId), 0, 0).Err()
	if err != nil {
		return 0, fmt.Errorf("unable to set balance_in_cents in cache for last account creation, err: %v", err)
	}

	err = m.Cache.Set(ctx, fmt.Sprintf("accounts/controlling_bank/%v", accountId), account.ControllingBank, 0).Err()
	if err != nil {
		return 0, fmt.Errorf("unable to set balance_in_cents in cache for last account creation, err: %v", err)
	}

	return accountId, err
}

func (m AccountModel) Get(id int64) (*Account, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, controlling_bank, metadata, frozen, version
        FROM accounts
        WHERE id = ?`

	var account Account

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, id).Scan(
		&account.Id,
		&account.ControllingBank,
		&account.Metadata,
		&account.Frozen,
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

	balanceInCents, err := m.Cache.Get(ctx, fmt.Sprintf("accounts/%v", account.Id)).Result()
	if err != nil {
		return nil, err
	}

	account.BalanceInCents, err = strconv.ParseInt(balanceInCents, 10, 64)
	if err != nil {
		return nil, err
	}

	return &account, nil
}

func (m AccountModel) Update(account *Account) error {
	query := `
        UPDATE accounts
        SET frozen = ?, metadata = ?, version = version + 1
        WHERE id = ? and controlling_bank = ? and version = ?`

	args := []interface{}{
		account.Frozen,
		account.Metadata,
		account.Id,
		account.ControllingBank,
		account.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.Db.QueryContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	err = m.Cache.Set(ctx, fmt.Sprintf("accounts/%v", account.Id), account.BalanceInCents, 0).Err()
	if err != nil {
		return fmt.Errorf("unable to set balance_in_cents in cache for last account creation, err: %v", err)
	}

	return nil
}
