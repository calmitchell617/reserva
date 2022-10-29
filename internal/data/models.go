package data

import (
	"database/sql"
	"errors"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Tokens   TokenModel
	Banks    BankModel
	Accounts AccountModel
	Cards    CardModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Tokens:   TokenModel{Db: db},
		Banks:    BankModel{Db: db},
		Accounts: AccountModel{Db: db},
		Cards:    CardModel{Db: db},
	}
}
