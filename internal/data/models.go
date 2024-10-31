package data

import (
	"database/sql"
	"errors"
	"time"
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

func NewModels(writeDb *sql.DB, readDb *sql.DB, queryTimeout time.Duration) Models {
	return Models{
		Accounts: AccountModel{
			WriteDB:      writeDb,
			ReadDb:       readDb,
			QueryTimeout: queryTimeout,
		},
		Cards: CardModel{
			WriteDb:      writeDb,
			ReadDb:       readDb,
			QueryTimeout: queryTimeout,
		},
		Transfers: TransferModel{
			WriteDb:      writeDb,
			ReadDb:       readDb,
			QueryTimeout: queryTimeout,
		},
		Users: UserModel{
			WriteDb:      writeDb,
			ReadDb:       readDb,
			QueryTimeout: queryTimeout,
		},
	}
}
