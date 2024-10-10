package data

import (
	"context"
	"database/sql"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
)

type AccountHolder struct {
	ID         int64  `json:"id"`
	ExternalID string `json:"external_id"`
}

func ValidateAccountHolder(v *validator.Validator, accountHolder *AccountHolder) {
	v.Check(accountHolder.ExternalID != "", "external_id", "must be provided")
}

type AccountHolderModel struct {
	DB *sql.DB
}

func (m AccountHolderModel) Insert(accountHolder *AccountHolder) error {
	query := `
        INSERT INTO account_holders (external_id)
		VALUES ($1)
        RETURNING id`

	args := []any{accountHolder.ExternalID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&accountHolder.ID)
}
