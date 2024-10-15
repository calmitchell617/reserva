package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Transfer struct {
	ID             int64     `json:"id"`
	CardID         int64     `json:"card_id"`
	FromAccountID  int64     `json:"from_account_id"`
	ToAccountID    int64     `json:"to_account_id"`
	RequestingUser User      `json:"requesting_user"`
	Amount         int64     `json:"amount"`
	CreatedAt      time.Time `json:"created_at"`
}

type TransferModel struct {
	DB *sql.DB
}

func (m *TransferModel) TransferFunds(transfer *Transfer, engine string) error {
	switch engine {
	case "postgresql":
		return m.TransferFundsPostgreSQL(transfer)
	}
	return fmt.Errorf("unsupported database engine")
}

func (m *TransferModel) TransferFundsPostgreSQL(transfer *Transfer) error {
	query := `
        SELECT transfer_funds($1, $2, $3, $4, $5, $6)
    `

	args := []interface{}{
		transfer.CardID,
		transfer.FromAccountID,
		transfer.ToAccountID,
		transfer.RequestingUser.ID,
		transfer.Amount,
		transfer.CreatedAt,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var transferID int64
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&transferID)
	if err != nil {
		fmt.Printf("Error transferring funds -> %v\n", err)
		return err
	}

	return nil
}
