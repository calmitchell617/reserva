package data

import (
	"database/sql"
	"errors"

	"github.com/go-redis/redis/v8"
)

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrEditConflict   = errors.New("edit conflict")
)

type Models struct {
	Tokens    TokenModel
	Banks     BankModel
	Accounts  AccountModel
	Cards     CardModel
	Transfers TransferModel
}

func NewModels(db *sql.DB, cache *redis.Client) Models {
	return Models{
		Tokens:    TokenModel{Db: db, Cache: cache},
		Banks:     BankModel{Db: db, Cache: cache},
		Accounts:  AccountModel{Db: db},
		Transfers: TransferModel{Db: db},
		Cards:     CardModel{Db: db},
	}
}
