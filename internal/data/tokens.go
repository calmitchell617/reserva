package data

import (
	"database/sql"
	"time"
)

type Token struct {
	Hash         []byte    `json:"-"`
	PermissionID int64     `json:"permission_id"`
	UserID       int64     `json:"-"`
	ExpiresAt    time.Time `json:"expiry"`
}

type TokenModel struct {
	DB *sql.DB
}
