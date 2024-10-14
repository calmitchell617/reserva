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
	// Movies           MovieModel
	Cards            CardModel
	TransferRequests TransferRequestModel
	Transfers        TransferModel
	Users            UserModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		// Movies:           MovieModel{DB: db},
		Cards:            CardModel{DB: db},
		TransferRequests: TransferRequestModel{DB: db},
		Transfers:        TransferModel{DB: db},
		Users:            UserModel{DB: db},
	}
}
