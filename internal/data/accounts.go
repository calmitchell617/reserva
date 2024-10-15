package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Account struct {
	ID             int64 `json:"id"`
	OrganizationID int64 `json:"organization_id"`
	Balance        int64 `json:"balance"`
	Frozen         bool  `json:"frozen"`
}

type AccountModel struct {
	DB *sql.DB
}

func (m AccountModel) Get(id int64, organizationId int64) (*Account, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT id, organization_id, balance, frozen
        FROM accounts
        WHERE id = $1
		and organization_id = $2`

	var account Account

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, id, organizationId).Scan(
		&account.ID,
		&account.OrganizationID,
		&account.Balance,
		&account.Frozen,
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
func (m AccountModel) GetFromCard(card *Card, engine string) (*Account, *Card, error) {
	switch engine {
	case "postgresql":
		return m.GetFromCardPostgreSQL(card)
	}
	return nil, nil, errors.New("unsupported database engine")
}

func (m AccountModel) GetFromCardPostgreSQL(card *Card) (*Account, *Card, error) {
	query := `
	SELECT accounts.id, accounts.organization_id, accounts.balance, accounts.frozen, cards.account_id, cards.expiration_date, cards.security_code, cards.frozen
	FROM accounts
	JOIN cards ON accounts.id = cards.account_id
	WHERE cards.id = $1
	`

	var account Account

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, card.ID).Scan(
		&account.ID,
		&account.OrganizationID,
		&account.Balance,
		&account.Frozen,
		&card.AccountID,
		&card.ExpirationDate,
		&card.SecurityCode,
		&card.Frozen,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil, ErrRecordNotFound
		default:
			return nil, nil, err
		}
	}

	return &account, card, nil
}

func (m AccountModel) UpdateBalance(tx *sql.Tx, accountID int64, amount int64) error {
	query := `
	UPDATE accounts
	SET balance = balance + $1
	WHERE id = $2
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := tx.ExecContext(ctx, query, amount, accountID)
	if err != nil {
		return err
	}

	return nil
}
