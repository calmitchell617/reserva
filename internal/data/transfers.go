package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
	"github.com/go-redis/redis/v8"
)

var (
	ErrNoPermission     = errors.New("you do not have permission to make a transfer from this account")
	ErrInsufficentFunds = errors.New("there are not enough funds to complete this tranactions")
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
	Db    *sql.DB
	Cache *redis.Client
}

func (m TransferModel) Insert(transfer *Transfer, requestingBank Bank) (*Transfer, error) {
	// Creates a new transfer
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// admin central bank can make a transfer from anywhere
	if !requestingBank.Admin {
		// if not admin, transfer can only be initiated by controlling bank of source account
		sourceControllingBank, err := m.Cache.Get(ctx, fmt.Sprintf("accounts/controlling_bank/%v", transfer.SourceAccountId)).Result()
		if err != nil {
			return nil, fmt.Errorf("unable to get source controlling bank, err: %v", err)
		}

		if sourceControllingBank != requestingBank.Username {
			return nil, ErrNoPermission
		}
	}

	// check sufficient funds
	sourceBalanceInCentsString, err := m.Cache.Get(ctx, fmt.Sprintf("accounts/%v", transfer.SourceAccountId)).Result()
	if err != nil {
		return nil, fmt.Errorf("unable to check source balance_in_cents in cache, err: %v", err)
	}

	sourceBalanceInCents, err := strconv.ParseInt(sourceBalanceInCentsString, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("unable to parse balance_in_cents int, err: %v", err)
	}

	if sourceBalanceInCents < transfer.AmountInCents {
		return nil, ErrInsufficentFunds
	}

	// insert into Singlestore
	query := `
        INSERT INTO transfers (source_account_id, target_account_id, amount_in_cents, created_at)
        VALUES (?, ?, ?, ?)`

	args := []interface{}{transfer.SourceAccountId, transfer.TargetAccountId, transfer.AmountInCents, transfer.CreatedAt}

	result, err := m.Db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("unable to run transfer insertion query, err: %v", err)
	}

	// change account transfers in cache
	err = m.Cache.DecrBy(ctx, fmt.Sprintf("accounts/%v", transfer.SourceAccountId), transfer.AmountInCents).Err()
	if err != nil {
		return nil, fmt.Errorf("unable to increment source account balance, err: %v", err)
	}

	err = m.Cache.IncrBy(ctx, fmt.Sprintf("accounts/%v", transfer.TargetAccountId), transfer.AmountInCents).Err()
	if err != nil {
		return nil, fmt.Errorf("unable to increment target account balance, err: %v", err)
	}

	transfer.Id, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return transfer, err
}
