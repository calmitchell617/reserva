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
	Movies         MovieModel
	Accounts       AccountModel
	AccountHolders AccountHolderModel
	Tokens         TokenModel
	Users          UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:         MovieModel{DB: db},
		Accounts:       AccountModel{DB: db},
		AccountHolders: AccountHolderModel{DB: db},
		Tokens:         TokenModel{DB: db},
		Users:          UserModel{DB: db},
	}
}
