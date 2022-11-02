package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"fmt"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
	"github.com/go-redis/redis/v8"
)

const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
	ScopePasswordReset  = "password-reset"
)

type Token struct {
	Plaintext    string    `json:"token"`
	Hash         []byte    `json:"-"`
	BankUsername string    `json:"-"`
	Expiry       time.Time `json:"expiry"`
	Admin        bool      `json:"-"`
}

func generateToken(bank *Bank, ttl time.Duration) (*Token, error) {
	token := &Token{
		BankUsername: bank.Username,
		Expiry:       time.Now().Add(ttl),
		Admin:        bank.Admin,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	Db    *sql.DB
	Cache *redis.Client
}

func (m TokenModel) New(bank *Bank, ttl time.Duration) (*Token, error) {
	token, err := generateToken(bank, ttl)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m TokenModel) Insert(token *Token) error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	tokenKey := fmt.Sprintf("tokens/%v", string(token.Hash))

	err := m.Cache.HSet(
		ctx,
		tokenKey,
		map[string]interface{}{
			"username": token.BankUsername,
			"expiry":   token.Expiry.Format(time.RFC3339),
			"admin":    token.Admin,
		},
	).Err()
	if err != nil {
		return fmt.Errorf("unable to insert into token cache, err: %v", err)
	}

	err = m.Cache.ExpireAt(ctx, tokenKey, token.Expiry).Err()
	if err != nil {
		return fmt.Errorf("unable to set token expiry, err: %v", err)
	}
	return nil
}

// func (m TokenModel) DeleteAllForBank(bankUsername string) error {
// 	query := `
//         DELETE FROM tokens
//         WHERE scope = $1 AND bank_id = $2`

// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	_, err := m.Db.ExecContext(ctx, query, scope, bankID)
// 	return err
// }

// func (m TokenModel) CheckAdminToken(token string) (bool, error) {
// 	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
// 	defer cancel()

// 	adminToken, err := m.Cache.Get(ctx, "admin-token").Result()
// 	if err != nil {
// 		return false, err
// 	}

// 	return adminToken == fmt.Sprint(sha256.Sum256([]byte(token))), nil
// }
