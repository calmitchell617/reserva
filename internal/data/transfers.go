package data

import (
	"context"
	"database/sql"
	"time"
)

type Transfer struct {
	ID                int64     `json:"id"`
	TransferRequestID int64     `json:"transfer_request_id"`
	FromAccountID     int64     `json:"from_account_id"`
	ToAccountID       int64     `json:"to_account_id"`
	Amount            int64     `json:"amount"`
	CreatedAt         time.Time `json:"created_at"`
}

type TransferModel struct {
	DB *sql.DB
}

func (m TransferModel) Insert(transfer *Transfer) error {
	query := `
        INSERT INTO transfers (transfer_request_id, from_account_id, to_account_id, amount, created_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	args := []any{transfer.TransferRequestID, transfer.FromAccountID, transfer.ToAccountID, transfer.Amount, transfer.CreatedAt}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return m.DB.QueryRowContext(ctx, query, args...).Scan(&transfer.ID)
}
