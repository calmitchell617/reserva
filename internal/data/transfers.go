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

func (m *TransferModel) TransferFunds(transfer *Transfer, engine string) (*Transfer, error) {
	switch engine {
	case "postgresql":
		return m.TransferFundsPostgreSQL(transfer)
	case "mariadb", "mysql":
		return m.TransferFundsMySQL(transfer)
	}
	return nil, fmt.Errorf("unsupported database engine")
}

func (m *TransferModel) TransferFundsMySQL(transfer *Transfer) (*Transfer, error) {
	query := `CALL transfer_funds(?, ?, ?, ?, ?, ?);`

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

	// Execute the stored procedure
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&transfer.ID)
	if err != nil {
		fmt.Printf("error transferring funds -> %v\n", err)
		return nil, err
	}

	// Check the transfer ID result
	if transfer.ID == -1 {
		fmt.Printf("transfer failed\n")
		return nil, fmt.Errorf("transfer failed")
	}

	return transfer, nil
}

func (m *TransferModel) TransferFundsPostgreSQL(transfer *Transfer) (*Transfer, error) {
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
		return nil, err
	}

	return transfer, nil
}

func (m *TransferModel) Delete(transferId int64, engine string) error {
	switch engine {
	case "postgresql":
		return m.DeletePostgreSQL(transferId)
	case "mariadb", "mysql":
		return m.DeleteMySQL(transferId)
	}
	return fmt.Errorf("unsupported database engine")
}

func (m *TransferModel) DeletePostgreSQL(transferId int64) error {
	query := `
		DELETE FROM transfers
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, transferId)
	if err != nil {
		fmt.Printf("Error deleting transfer -> %v\n", err)
		return err
	}

	return nil
}

func (m *TransferModel) DeleteMySQL(transferId int64) error {
	query := `
		DELETE FROM transfers
		WHERE id = ?
	`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, query, transferId)
	if err != nil {
		fmt.Printf("Error deleting transfer -> %v\n", err)
		return err
	}

	return nil
}
