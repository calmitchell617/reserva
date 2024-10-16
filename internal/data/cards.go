package data

import (
	"database/sql"
	"time"
)

type Card struct {
	ID             int64     `json:"id"`
	AccountID      int64     `json:"account_id"`
	ExpirationDate time.Time `json:"expiration_date"`
	SecurityCode   int       `json:"security_code"`
	Frozen         bool      `json:"frozen"`
}

type CardModel struct {
	DB *sql.DB
}
