package data

import (
	"context"
	"crypto/ed25519"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/calmitchell617/reserva/internal/validator"
)

var (
	ErrDuplicateCardId = errors.New("duplicate card number")
)

type Card struct {
	Id         int64              `json:"id"`
	AccountId  int64              `json:"account_id"`
	PrivateKey ed25519.PrivateKey `json:"-"`
	Password   password           `json:"-"`
	Expiry     time.Time          `json:"expiry"`
	Version    int64              `json:"version"`
}

func ValidateCard(v *validator.Validator, card *Card) {
	if card.Password.plaintext != nil {
		ValidatePasswordPlaintext(v, *card.Password.plaintext)
	}

	if card.Password.hash == nil {
		panic("missing password hash for card")
	}
}

type CardModel struct {
	Db *sql.DB
}

func (m CardModel) Insert(card *Card, bankId int64) error {
	query := `
				insert into cards (id, account_id, private_key, password_hash, expiry)
				select $1, $2, $3, $4, $5 from accounts where bank_id = $6
        RETURNING id, version`

	args := []interface{}{card.Id, card.AccountId, card.PrivateKey, card.Password.hash, card.Expiry, bankId}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, args...).Scan(&card.Id, &card.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "cards_id_key"`:
			return ErrDuplicateCardId
		default:
			return err
		}
	}
	return nil
}

func (m CardModel) Get(id int64, bankId int64) (*Card, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}

	query := `
        SELECT cards.id, cards.account_id, cards.private_key, cards.password_hash, cards.expiry, cards.version
        FROM cards
				inner join accounts on cards.account_id = accounts.id
        WHERE cards.id = $1 and accounts.bank_id = $2`

	var card Card

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, id, bankId).Scan(
		&card.Id,
		&card.AccountId,
		&card.PrivateKey,
		&card.Password.hash,
		&card.Version,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &card, nil
}

func (m CardModel) Update(card *Card, bankId int64) error {
	query := `
        UPDATE cards
        SET cards.password_hash = $1, cards.version = cards.version + 1
				from accounts
				where cards.account_id = accounts.id
        WHERE cards.id = $2 and accounts.bank_id = $3 and cards.version = $4
        RETURNING cards.version`

	args := []interface{}{
		card.Password.hash,
		card.Id,
		bankId,
		card.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := m.Db.QueryRowContext(ctx, query, args...).Scan(&card.Version)
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

func (m CardModel) Delete(id int64, bankId int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	query := `
        DELETE FROM cards
				using accounts
				where cards.account_id = accounts.id
        WHERE cards.id = $1 and accounts.bank_id = $2`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	result, err := m.Db.ExecContext(ctx, query, id, bankId)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (m CardModel) GetAll(bankId int64, filters Filters) ([]*Card, Metadata, error) {
	query := fmt.Sprintf(`
        SELECT count(*) OVER(), id, account_id, private_key, password_hash, expiry, version
        FROM cards
				inner join accounts on cards.account_id = accounts.id
        where bank_id = $1
				and accounts.bank_id = $2
        ORDER BY %s %s, id ASC
        LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{bankId, filters.limit(), filters.offset()}

	rows, err := m.Db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err
	}

	defer rows.Close()

	totalRecords := 0
	cards := []*Card{}

	for rows.Next() {
		var card Card

		err := rows.Scan(
			&totalRecords,
			&card.Id,
			&card.AccountId,
			&card.PrivateKey,
			&card.Password.hash,
			&card.Expiry,
			&card.Version,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		cards = append(cards, &card)
	}

	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return cards, metadata, nil
}
