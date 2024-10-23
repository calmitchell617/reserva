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

func NewModels(writeDb *sql.DB, readDb *sql.DB) Models {
	return Models{
		Accounts: AccountModel{
			WriteDB: writeDb,
			ReadDb:  readDb,
		},
		Cards: CardModel{
			WriteDb: writeDb,
			ReadDb:  readDb,
		},
		Transfers: TransferModel{
			WriteDb: writeDb,
			ReadDb:  readDb,
		},
		Users: UserModel{
			WriteDb: writeDb,
			ReadDb:  readDb,
		},
	}
}
