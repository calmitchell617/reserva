package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
	"github.com/calmitchell617/reserva/pkg"
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
	Db *sql.DB
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
	return accountId, err
}

func (m AccountModel) Get(id int64) (*Account, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, controlling_bank, metadata, balance_in_cents, frozen, version
        FROM accounts
        WHERE id = ?`

	var account Account

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, id).Scan(
		&account.Id,
		&account.ControllingBank,
		&account.Metadata,
		&account.BalanceInCents,
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

	return &account, nil
}

func (m AccountModel) Update(account *Account) error {
	query := `
        UPDATE accounts
        SET balance_in_cents = ?, frozen = ?, metadata = ?, version = version + 1
        WHERE id = ? and controlling_bank = ? and version = ?`

	args := []interface{}{
		account.BalanceInCents,
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

	return nil
}

// func (m AccountModel) Delete(id int64, bankId int64) error {
// 	if id < 1 {
// 		return ErrRecordNotFound
// 	}

// 	query := `
//         DELETE FROM accounts
//         WHERE id = $1 and bank_id = $2`

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	result, err := m.Db.ExecContext(ctx, query, id, bankId)
// 	if err != nil {
// 		return err
// 	}

// 	rowsAffected, err := result.RowsAffected()
// 	if err != nil {
// 		return err
// 	}

// 	if rowsAffected == 0 {
// 		return ErrRecordNotFound
// 	}

// 	return nil
// }

// func (m AccountModel) GetAll(bankId int64, filters Filters) ([]*Account, Metadata, error) {
// 	query := fmt.Sprintf(`
//         SELECT count(*) OVER(), id, bank_id, balance_in_cents, frozen, version
//         FROM accounts
//         where bank_id = $1
//         ORDER BY %s %s, id ASC
//         LIMIT $2 OFFSET $3`, filters.sortColumn(), filters.sortDirection())

// 	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
// 	defer cancel()

// 	args := []interface{}{bankId, filters.limit(), filters.offset()}

// 	rows, err := m.Db.QueryContext(ctx, query, args...)
// 	if err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	defer rows.Close()

// 	totalRecords := 0
// 	accounts := []*Account{}

// 	for rows.Next() {
// 		var account Account

// 		err := rows.Scan(
// 			&totalRecords,
// 			&account.Id,
// 			&account.BankId,
// 			&account.BalanceInCents,
// 			&account.Frozen,
// 			&account.Version,
// 		)
// 		if err != nil {
// 			return nil, Metadata{}, err
// 		}

// 		accounts = append(accounts, &account)
// 	}

// 	if err = rows.Err(); err != nil {
// 		return nil, Metadata{}, err
// 	}

// 	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

// 	return accounts, metadata, nil
// }
