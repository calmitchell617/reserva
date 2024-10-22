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
	Accounts  AccountModel
	Cards     CardModel
	Transfers TransferModel
	Users     UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Accounts:  AccountModel{DB: db},
		Cards:     CardModel{DB: db},
		Transfers: TransferModel{DB: db},
		Users:     UserModel{DB: db},
	}
}
