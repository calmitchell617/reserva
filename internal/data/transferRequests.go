package data

import (
	"context"
	"database/sql"
	"time"
)

type TransferRequest struct {
	ID                 int64     `json:"id"`
	CardId             int64     `json:"card_id"`
	IssuingAccountID   int64     `json:"issuing_account_id"`
	AcquiringAccountID int64     `json:"acquiring_account_id"`
	Amount             int64     `json:"amount"`
	CreatedAt          time.Time `json:"created_at"`
}

type TransferRequestModel struct {
	DB *sql.DB
}

func (m TransferRequestModel) Insert(transferRequest *TransferRequest) (*TransferRequest, error) {
	query := `
        INSERT INTO transfer_requests (card_id, issuing_account_id, acquiring_account_id, amount, created_at) 
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	args := []any{transferRequest.CardId, transferRequest.IssuingAccountID, transferRequest.AcquiringAccountID, transferRequest.Amount, transferRequest.CreatedAt}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	m.DB.QueryRowContext(ctx, query, args...).Scan(&transferRequest.ID)

	return transferRequest, nil
}
