package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
)

type Transfer struct {
	Id              int64     `json:"id"`
	SourceAccountId int64     `json:"source_account_id"`
	TargetAccountId int64     `json:"target_account_id"`
	CreatedAt       time.Time `json:"created_at"`
	AmountInCents   int64     `json:"amount_in_cents"`
}

func ValidateTransfer(v *validator.Validator, transfer *Transfer) {
	v.Check(transfer.SourceAccountId != 0, "source_account_id", "must be provided")
	v.Check(transfer.TargetAccountId != 0, "target_account_id", "must be provided")
	v.Check(transfer.AmountInCents != 0, "amount_in_cents", "must be provided")
	v.Check(transfer.AmountInCents > 0, "amount_in_cents", "must be greater than 0")
}

type TransferModel struct {
	Db *sql.DB
}

func (m TransferModel) Insert(transfer *Transfer) (*Transfer, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := m.Db.BeginTx(ctx, &sql.TxOptions{})
	defer tx.Rollback()
	if err != nil {
		return nil, err
	}

	query := `
        INSERT INTO transfers (source_account_id, target_account_id, amount_in_cents, created_at)
        VALUES (?, ?, ?, ?)`

	args := []interface{}{transfer.SourceAccountId, transfer.TargetAccountId, transfer.AmountInCents, transfer.CreatedAt}

	result, err := tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to run transfer insertion query, err: %v", err)
	}

	transfer.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	query = `update accounts set balance_in_cents = balance_in_cents - ?, version = version + 1 where id = ?`

	args = []interface{}{transfer.AmountInCents, transfer.SourceAccountId}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to run transfer source account update query, err: %v", err)
	}

	query = `update accounts set balance_in_cents = balance_in_cents + ?, version = version + 1 where id = ?`

	args = []interface{}{transfer.AmountInCents, transfer.TargetAccountId}

	_, err = tx.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to run transfer target account update query, err: %v", err)
	}

	err = tx.Commit()
	if err != nil {
		return transfer, fmt.Errorf("unable to commit transfer transaction, err: %v", err)
	}

	return transfer, err
}
