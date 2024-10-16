package data

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

var AnonymousUser = &User{}

type User struct {
	ID             int64 `json:"id"`
	OrganizationID int64 `json:"organization_id"`
	Frozen         bool  `json:"frozen"`
	AccountID      int64 `json:"account_id"`
	Card           Card  `json:"card"`
	Token          Token `json:"token"`
}

func (u *User) IsAnonymous() bool {
	return u == AnonymousUser
}

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) GetAll(engine string) ([]*User, error) {
	switch engine {
	case "postgresql":
		return m.GetAllUsersPostgreSQL()
	case "mariadb":
		return m.GetAllUsersMariaDB()
	}
	return nil, fmt.Errorf("unsupported database engine")
}

func (m UserModel) GetAllUsersPostgreSQL() ([]*User, error) {
	query := `
SELECT
	USERS.ID,
	USERS.ORGANIZATION_ID AS ORGANIZATION_ID,
	USERS.FROZEN,
	ACCOUNTS.ID AS ACCOUNT_ID,
	CARDS.ID AS CARD_ID,
	CARDS.Expiration_Date as Expiration_Date,
	CARDS.SECURITY_CODE as security_code,
	CARDS.FROZEN as card_frozen,
	tokens.HASH 
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error running query: %v", err)
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {

		var user User
		var card Card
		var token Token

		err := rows.Scan(
			&user.ID,
			&user.OrganizationID,
			&user.Frozen,
			&user.AccountID,
			&card.ID,
			&card.ExpirationDate,
			&card.SecurityCode,
			&card.Frozen,
			&token.Hash,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		user.Card = card
		user.Token = token

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return users, nil
}

func (m UserModel) GetAllUsersMariaDB() ([]*User, error) {
	query := `
SELECT
	users.ID,
	users.ORGANIZATION_ID AS ORGANIZATION_ID,
	users.FROZEN,
	accounts.ID AS ACCOUNT_ID,
	cards.ID AS CARD_ID,
	cards.Expiration_Date as Expiration_Date,
	cards.SECURITY_CODE as security_code,
	cards.FROZEN as card_frozen,
	tokens.HASH 
FROM
	users
	JOIN organizations ON users.ORGANIZATION_ID = organizations.ID
	JOIN accounts ON organizations.ID = accounts.ORGANIZATION_ID
	JOIN tokens ON users.ID = tokens.USER_ID
	JOIN cards ON accounts.ID = cards.ACCOUNT_ID
ORDER BY
	users.ID,
	ORGANIZATION_ID,
	ACCOUNT_ID`

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("error running query: %v", err)
	}
	defer rows.Close()

	users := []*User{}

	for rows.Next() {
		var user User
		var card Card
		var token Token

		err := rows.Scan(
			&user.ID,
			&user.OrganizationID,
			&user.Frozen,
			&user.AccountID,
			&card.ID,
			&card.ExpirationDate,
			&card.SecurityCode,
			&card.Frozen,
			&token.Hash,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning row: %v", err)
		}

		user.Card = card
		user.Token = token

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	return users, nil
}

func (m UserModel) GetForToken(tokenHash []byte, engine string) ([]*User, error) {
	switch engine {
	case "postgresql":
		return m.GetForTokenPostgreSQL(tokenHash)
	case "mariadb":
		return m.GetForTokenMariaDB(tokenHash)
	}
	return nil, fmt.Errorf("unsupported database engine")
}

func (m UserModel) GetForTokenMariaDB(tokenHash []byte) ([]*User, error) {
	query := `
        SELECT users.id, users.organization_id, users.frozen, tokens.hash, tokens.permission_id, tokens.expires_at
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = ?`

	var users []*User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, tokenHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var user User
		var token Token

		err := rows.Scan(
			&user.ID,
			&user.OrganizationID,
			&user.Frozen,
			&token.Hash,
			&token.PermissionID,
			&token.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		user.Token = token

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (m UserModel) GetForTokenPostgreSQL(tokenHash []byte) ([]*User, error) {
	query := `
        SELECT users.id, users.organization_id, users.frozen, tokens.hash, tokens.permission_id, tokens.expires_at
        FROM users
        INNER JOIN tokens
        ON users.id = tokens.user_id
        WHERE tokens.hash = $1`

	var users []*User

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query, tokenHash)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {

		var user User
		var token Token

		err := rows.Scan(
			&user.ID,
			&user.OrganizationID,
			&user.Frozen,
			&token.Hash,
			&token.PermissionID,
			&token.ExpiresAt,
		)
		if err != nil {
			return nil, err
		}

		user.Token = token

		users = append(users, &user)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}
