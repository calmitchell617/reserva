package data

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/calmitchell617/reserva/internal/validator"
	"github.com/calmitchell617/reserva/pkg"
	"github.com/go-redis/redis/v8"
)

var AnonymousBank = &Bank{}
var Admin = &Bank{}

var (
	ErrDuplicateUsername = errors.New("duplicate username")
)

type Bank struct {
	Username string   `json:"username"`
	Admin    bool     `json:"admin"`
	Password password `json:"-"`
	Version  int64    `json:"-"`
}

func (u *Bank) IsAnonymous() bool {
	return u == AnonymousBank
}

func ValidatePasswordPlaintext(v *validator.Validator, password string) {
	v.Check(password != "", "password", "must be provided")
	v.Check(utf8.RuneCountInString(password) >= 8, "password", "must be at least 8 characters long")
	v.Check(utf8.RuneCountInString(password) <= 72, "password", "must not be more than 72 characters long")
}

func ValidateBank(v *validator.Validator, bank *Bank) {
	v.Check(bank.Username != "", "username", "must be provided")
	v.Check(utf8.RuneCountInString(bank.Username) <= 32, "username", "must not be more than 32 characters long")
	v.Check(pkg.IsAlphanumeric(bank.Username), "username", "must only contain alphanumeric characters - no spaces, special characters, or numbers")

	if bank.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *bank.Password.plaintext)
	}

	if bank.Password.hash == nil {
		panic("missing password hash for bank")
	}
}

type BankModel struct {
	Db    *sql.DB
	Cache *redis.Client
}

func (m BankModel) Insert(bank *Bank) error {
	query := `
        INSERT INTO banks (username, admin, password_hash) 
        VALUES (?, ?, ?)`

	args := []interface{}{bank.Username, bank.Admin, bank.Password.hash}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.Db.ExecContext(ctx, query, args...)
	if err != nil {
		switch {
		case strings.HasPrefix(err.Error(), "Error 1062"):
			return ErrDuplicateUsername
		default:
			return err
		}
	}

	return nil
}

func (m BankModel) GetByUsername(username string) (*Bank, error) {
	query := `
        SELECT
					username,
					admin,
					password_hash,
					version
        FROM banks
        WHERE username = ?`

	var bank Bank

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, username).Scan(
		&bank.Username,
		&bank.Admin,
		&bank.Password.hash,
		&bank.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &bank, nil
}

func (m BankModel) Update(bank *Bank) error {
	query := `
        UPDATE banks 
        SET
					admin = ?,
					password_hash = ?,
					version = version + 1
        WHERE username = ? AND version = ?`

	args := []interface{}{
		bank.Admin,
		bank.Password.hash,
		bank.Username,
		bank.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := m.Db.QueryContext(ctx, query, args...)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}

func (m BankModel) GetByToken(tokenScope, tokenPlaintext string) (*Bank, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	token := tokenHash[:]
	tokenHashString := fmt.Sprintf("tokens/%v", string(token))

	bankMap, err := m.Cache.HGetAll(ctx, tokenHashString).Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	username, ok := bankMap["username"]
	if !ok {
		return nil, fmt.Errorf("bank username not stored in token hash")
	}

	adminString, ok := bankMap["admin"]
	if !ok {
		return nil, errors.New("bank admin status not stored in token hash")
	}

	admin, err := strconv.ParseBool(adminString)
	if err != nil {
		return nil, errors.New("unable to parse admin bool from cache token")
	}

	expiryString, ok := bankMap["expiry"]
	if !ok {
		return nil, fmt.Errorf("expiry not stored in token hash")
	}

	expiry, err := time.Parse(time.RFC3339, expiryString)
	if err != nil {
		return nil, errors.New("unable to parse expiry time from cache token")
	}

	// check if it expired
	if expiry.Before(time.Now()) {
		return nil, ErrRecordNotFound
	}

	bank := Bank{
		Username: username,
		Admin:    admin,
	}

	query := `
		SELECT 
			username,
			admin,
			password_hash,
			version
		FROM banks
		WHERE username = ?`

	args := []interface{}{bank.Username}

	err = m.Db.QueryRowContext(ctx, query, args...).Scan(
		&bank.Username,
		&bank.Admin,
		&bank.Password.hash,
		&bank.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &bank, nil
}

func (m BankModel) GetByTokenForAuthentication(tokenScope, tokenPlaintext string) (*Bank, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	token := tokenHash[:]
	tokenHashString := fmt.Sprintf("tokens/%v", string(token))

	bankMap, err := m.Cache.HGetAll(ctx, tokenHashString).Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	username, ok := bankMap["username"]
	if !ok {
		return nil, fmt.Errorf("bank username not stored in token hash")
	}

	adminString, ok := bankMap["admin"]
	if !ok {
		return nil, errors.New("bank admin status not stored in token hash")
	}

	admin, err := strconv.ParseBool(adminString)
	if err != nil {
		return nil, errors.New("unable to parse admin bool from cache token")
	}

	expiryString, ok := bankMap["expiry"]
	if !ok {
		return nil, fmt.Errorf("expiry not stored in token hash")
	}

	expiry, err := time.Parse(time.RFC3339, expiryString)
	if err != nil {
		return nil, errors.New("unable to parse expiry time from cache token")
	}

	// check if it expired
	if expiry.Before(time.Now()) {
		return nil, ErrRecordNotFound
	}

	bank := Bank{
		Username: username,
		Admin:    admin,
	}

	return &bank, nil
}
