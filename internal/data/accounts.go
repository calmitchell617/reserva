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
