package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/calmitchell617/reserva/internal/jsonlog"
)

var (
	host string
)

type AuthTokenResponse struct {
	AuthToken AuthToken `json:"authentication_token"`
}

type AuthToken struct {
	Token  string    `json:"token"`
	Expiry time.Time `json:"expiry"`
}

func main() {

	flag.StringVar(&host, "host", "", "reserva host")

	flag.Parse()

	if host == "" {
		fmt.Println("You must enter a host to load test")
		os.Exit(1)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	client := &http.Client{Timeout: time.Second * 3}

	// Get auth token for admin bank

	json_data, err := json.Marshal(map[string]string{"username": "adminBank", "password": "Mypass123"})
	if err != nil {
		logger.PrintFatal(errors.New("unable to marshal json data"), map[string]string{"error": err.Error()})
	}

	response, err := client.Post(fmt.Sprintf("%v/v1/tokens/authentication", host), "application/json; charset=UTF-8", bytes.NewBuffer(json_data))
	if err != nil {
		logger.PrintFatal(errors.New("unable to get initial admin bank auth token"), map[string]string{"error": err.Error()})
	}

	authTokenResp := AuthTokenResponse{}
	json.NewDecoder(response.Body).Decode(&authTokenResp)
	response.Body.Close()
	adminBearerToken := fmt.Sprintf("Bearer %v", authTokenResp.AuthToken.Token)

	bankBearerTokens := []string{}

	numBanks := 5

	// Create test banks
	for i := 0; i < numBanks; i++ {
		json_data, err := json.Marshal(map[string]any{"username": fmt.Sprintf("bank%v", i), "admin": false, "password": "Mypass123"})
		if err != nil {
			logger.PrintFatal(errors.New("unable to marshal new bank json data"), map[string]string{"error": err.Error()})
		}
		r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/banks", host), bytes.NewBuffer(json_data))
		if err != nil {
			logger.PrintFatal(errors.New("unable to create new bank request"), map[string]string{"error": err.Error()})
		}

		r.Header.Add("Authorization", adminBearerToken)

		_, err = client.Do(r)
		if err != nil {
			logger.PrintFatal(errors.New("unable to perform new bank request"), map[string]string{"error": err.Error()})
		}

		// get the auth token for that bank

		json_data, err = json.Marshal(map[string]string{"username": fmt.Sprintf("bank%v", i), "password": "Mypass123"})
		if err != nil {
			logger.PrintFatal(errors.New("unable to marshal json data"), map[string]string{"error": err.Error()})
		}

		response, err := client.Post(fmt.Sprintf("%v/v1/tokens/authentication", host), "application/json; charset=UTF-8", bytes.NewBuffer(json_data))
		if err != nil {
			logger.PrintFatal(errors.New("unable to get bank auth token"), map[string]string{"error": err.Error()})
		}

		authTokenResp := AuthTokenResponse{}
		json.NewDecoder(response.Body).Decode(&authTokenResp)
		response.Body.Close()
		bearerToken := fmt.Sprintf("Bearer %v", authTokenResp.AuthToken.Token)
		bankBearerTokens = append(bankBearerTokens, bearerToken)
	}

	// fmt.Println(bankBearerTokens)

	// create test accounts
	for i := 0; i < 1000; i++ {
		json_data, err := json.Marshal(map[string]any{"metadata": `{"mykey": "myval"}`})
		if err != nil {
			logger.PrintFatal(errors.New("unable to marshal new account json data"), map[string]string{"error": err.Error()})
		}
		r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/accounts", host), bytes.NewBuffer(json_data))
		if err != nil {
			logger.PrintFatal(errors.New("unable to create new account request"), map[string]string{"error": err.Error()})
		}

		randomBearerTokenIndex := rand.Intn(numBanks)

		r.Header.Add("Authorization", bankBearerTokens[randomBearerTokenIndex])

		_, err = client.Do(r)
		if err != nil {
			logger.PrintFatal(errors.New("unable to perform new account request"), map[string]string{"error": err.Error()})
		}

		// _, err = ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	logger.PrintFatal(errors.New("unable to read new account response"), map[string]string{"error": err.Error()})
		// }

		// change money supply by adding to account
		json_data, err = json.Marshal(map[string]any{"id": i + 1, "change_in_cents": 100000000})
		if err != nil {
			logger.PrintFatal(errors.New("unable to create change money supply json"), map[string]string{"error": err.Error()})
		}
		r, err = http.NewRequest("PATCH", fmt.Sprintf("%v/v1/accounts/change_money_supply", host), bytes.NewBuffer(json_data))
		if err != nil {
			logger.PrintFatal(errors.New("unable to create new money supply request"), map[string]string{"error": err.Error()})
		}

		r.Header.Add("Authorization", adminBearerToken)

		_, err = client.Do(r)
		if err != nil {
			logger.PrintFatal(errors.New("unable to perform money supply request"), map[string]string{"error": err.Error()})
		}

		// _, err = ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	logger.PrintFatal(errors.New("unable to read money supply response"), map[string]string{"error": err.Error()})
		// }

		// if i%1000 == 0 {
		// 	fmt.Println(string(body))
		// }
	}

	// // create test transactions
	for i := 0; i < 10000; i++ {
		sourceId := rand.Intn(1000)
		targetId := rand.Intn(1000)
		amountInCents := rand.Intn(10)
		json_data, err := json.Marshal(map[string]any{"source_account_id": sourceId, "target_account_id": targetId, "amount_in_cents": amountInCents})
		if err != nil {
			logger.PrintFatal(errors.New("unable to marshal new transfer json data"), map[string]string{"error": err.Error()})
		}
		r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/transfers", host), bytes.NewBuffer(json_data))
		if err != nil {
			logger.PrintFatal(errors.New("unable to create new transfer request"), map[string]string{"error": err.Error()})
		}

		r.Header.Add("Authorization", adminBearerToken)

		_, err = client.Do(r)
		if err != nil {
			logger.PrintFatal(errors.New("unable to perform new transfer request"), map[string]string{"error": err.Error()})
		}

		// _, err = ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	logger.PrintFatal(errors.New("unable to read transfer response"), map[string]string{"error": err.Error()})
		// }

		if i%1000 == 0 {
			fmt.Println(i)
		}
	}
}
