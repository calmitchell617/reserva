package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var AnonymousUser = &User{}

type User struct {
	ID             int64  `json:"id"`
	OrganizationID int64  `json:"organization_id"`
	Frozen         bool   `json:"frozen"`
	AccountID      int64  `json:"account_id"`
	CardID         int64  `json:"card_id"`
	TokenHash      []byte `json:"token_hash"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) GetAll() ([]*User, error) {
	query := `
SELECT
	USERS.ID,
	USERS.ORGANIZATION_ID AS ORGANIZATION_ID,
	USERS.FROZEN,
	ACCOUNTS.ID AS ACCOUNT_ID,
	CARDS.ID AS CARD_ID,
	HASH
FROM
	USERS
	JOIN ORGANIZATIONS ON USERS.ORGANIZATION_ID = ORGANIZATIONS.ID
	JOIN ACCOUNTS ON ORGANIZATIONS.ID = ACCOUNTS.ORGANIZATION_ID
	JOIN TOKENS ON USERS.ID = TOKENS.USER_ID
	JOIN CARDS ON ACCOUNTS.ID = CARDS.ACCOUNT_ID
ORDER BY
	USERS.ID,
	ORGANIZATION_ID,
	ACCOUNT_ID`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user User
		err := rows.Scan(
			&user.ID,
			&user.OrganizationID,
			&user.Frozen,
			&user.AccountID,
			&user.CardID,
			&user.TokenHash,
		)
		if err != nil {
			return nil, err
		}

		users = append(users, &user)
	}

	return users, nil
}

func (m UserModel) GetForToken(tokenHash []byte) (*User, error) {
	query := `
        SELECT users.id, users.organization_id, users.frozen
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1
        AND tokens.expires_at > $2`

	args := []any{tokenHash[:], time.Now()}

	var user User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, args...).Scan(
		&user.ID,
		&user.OrganizationID,
		&user.Frozen,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &user, nil
}
