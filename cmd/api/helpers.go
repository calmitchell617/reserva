package main

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

func (app *application) readUsernameParam(r *http.Request) (string, error) {
	// get username from route
	params := httprouter.ParamsFromContext(r.Context())

	username := params.ByName("username")
	if username == "" {
		return "", errors.New("username not present in query params")
	}
	return username, nil
}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	// get id from route
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

type envelope map[string]interface{}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	// respond with prettified JSON

	// indent
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		return err
	}

	// newlines
	js = append(js, '\n')

	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, dst interface{}) error {

	// set max incoming bytes
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)

	// only allow fields specified in Go's struct tags to be decoded
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

// A bunch of stuff to do with cards that I'd like to implement, but didn't have time

func generateCardNumber() (int64, error) {
	var cardNumber int64
	cardString := "29"

	for i := 0; i < 13; i++ {
		nBig, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return cardNumber, err
		}
		n := fmt.Sprint(nBig.Int64())

		cardString = cardString + n
	}

	baseNumber, err := strconv.ParseInt(cardString, 10, 64)
	if err != nil {
		return cardNumber, err
	}

	luhnNumber := getLuhn(baseNumber)

	cardString = cardString + fmt.Sprint(luhnNumber)

	cardNumber, err = strconv.ParseInt(cardString, 10, 64)
	if err != nil {
		return cardNumber, err
	}

	return cardNumber, err
}

func checksum(n int64) int64 {
	var l int64
	for i := 0; n > 0; i++ {
		c := n % 10
		if i%2 == 0 {
			c = c * 2
			if c > 9 {
				c = c%10 + c/10
			}
		}
		l += c
		n = n / 10
	}
	return l % 10
}

func getLuhn(n int64) int64 {
	// credit card numbers are usually "luhn" numbers, which are numbers that conform
	// to a certain checksum algorithm
	cn := checksum(n)

	if cn == 0 {
		return 0
	}
	return 10 - cn
}
