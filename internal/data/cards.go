package data

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type Card struct {
	ID             int64     `json:"id"`
	AccountID      int64     `json:"account_id"`
	ExpirationDate time.Time `json:"expiration_date"`
	Frozen         bool      `json:"frozen"`
}

type CardModel struct {
	DB *sql.DB
}

func (m CardModel) GetAll() ([]*Card, error) {
	query := `
		SELECT id, account_id, expiration_date, frozen
		FROM cards order by id`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := m.DB.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cards := []*Card{}

	for rows.Next() {
		var card Card
		err := rows.Scan(
			&card.ID,
			&card.AccountID,
			&card.ExpirationDate,
			&card.Frozen,
		)
		if err != nil {
			return nil, err
		}

		cards = append(cards, &card)
	}

	return cards, nil
}

func (m CardModel) GetAccountIDFromCard(cardId int64) (int64, error) {
	query := `SELECT account_id from cards where id = $1`

	var accountId int64

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.DB.QueryRowContext(ctx, query, cardId).Scan(&accountId)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return 0, ErrRecordNotFound
		default:
			return 0, err
		}
	}

	return accountId, nil
}
