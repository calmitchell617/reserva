package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/calmitchell617/reserva/internal/jsonlog"
	"golang.org/x/sync/errgroup"
)

var (
	host        string
	concurrency int
)

type AuthTokenResponse struct {
	AuthToken AuthToken `json:"authentication_token"`
}

type AuthToken struct {
	Token  string    `json:"token"`
	Expiry time.Time `json:"expiry"`
}

func main() {
	start := time.Now()

	flag.StringVar(&host, "host", "", "reserva host")
	flag.IntVar(&concurrency, "concurrency", 100, "max num requests")

	flag.Parse()

	if host == "" {
		fmt.Println("You must enter a host to load test")
		os.Exit(1)
	}

	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)
	client := &http.Client{Transport: &http.Transport{
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     concurrency,
	}}

	// Get auth token for admin bank

	json_data, err := json.Marshal(map[string]string{"username": "adminBank", "password": "Mypass123"})
	if err != nil {
		logger.PrintFatal(errors.New("unable to marshal json data"), map[string]string{"error": err.Error()})
	}

	resp, err := client.Post(fmt.Sprintf("%v/v1/tokens/authentication", host), "application/json; charset=UTF-8", bytes.NewBuffer(json_data))
	if err != nil {
		logger.PrintFatal(errors.New("unable to get initial admin bank auth token"), map[string]string{"error": err.Error()})
	}

	authTokenResp := AuthTokenResponse{}
	json.NewDecoder(resp.Body).Decode(&authTokenResp)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	adminBearerToken := fmt.Sprintf("Bearer %v", authTokenResp.AuthToken.Token)

	numBanks := 100
	bankBearerTokens := make([]string, numBanks)
	errGroup := new(errgroup.Group)
	errGroup.SetLimit(concurrency)

	fmt.Printf("finished setting up admin bank, took %v\n", time.Since(start))
	start = time.Now()

	// Create test banks
	for i := 0; i < numBanks; i++ {
		i := i
		errGroup.Go(func() error {
			json_data, err := json.Marshal(map[string]any{"username": fmt.Sprintf("bank%v", i), "admin": false, "password": "Mypass123"})
			if err != nil {
				return err
			}
			r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/banks", host), bytes.NewBuffer(json_data))
			if err != nil {
				return err
			}

			r.Header.Add("Authorization", adminBearerToken)

			resp, err := client.Do(r)
			if err != nil {
				return err
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()

			// get the auth token for that bank

			json_data, err = json.Marshal(map[string]string{"username": fmt.Sprintf("bank%v", i), "password": "Mypass123"})
			if err != nil {
				return err
			}

			response, err := client.Post(fmt.Sprintf("%v/v1/tokens/authentication", host), "application/json; charset=UTF-8", bytes.NewBuffer(json_data))
			if err != nil {
				return err
			}

			authTokenResp := AuthTokenResponse{}
			json.NewDecoder(response.Body).Decode(&authTokenResp)
			io.Copy(ioutil.Discard, resp.Body)
			response.Body.Close()
			bearerToken := fmt.Sprintf("%vBearer %v", i, authTokenResp.AuthToken.Token)
			bankBearerTokens[i] = bearerToken
			return nil
		})
	}
	err = errGroup.Wait()
	if err != nil {
		logger.PrintFatal(errors.New("unable to create banks"), map[string]string{"error": err.Error()})
	}
	fmt.Printf("done creating banks, took %v\n", time.Since(start))
	start = time.Now()

	// create test accounts
	for i := 0; i < 1000; i++ {
		i := i
		errGroup.Go(func() error {
			json_data, err := json.Marshal(map[string]any{"metadata": `{"mykey": "myval"}`})
			if err != nil {
				return err
			}
			r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/accounts", host), bytes.NewBuffer(json_data))
			if err != nil {
				return err
			}

			randomBearerTokenIndex := rand.Intn(numBanks)

			r.Header.Add("Authorization", bankBearerTokens[randomBearerTokenIndex])

			resp, err := client.Do(r)
			if err != nil {
				return err
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()

			// change money supply by adding to account
			json_data, err = json.Marshal(map[string]any{"id": i + 1, "change_in_cents": 100000000})
			if err != nil {
				return err
			}
			r, err = http.NewRequest("PATCH", fmt.Sprintf("%v/v1/accounts/change_money_supply", host), bytes.NewBuffer(json_data))
			if err != nil {
				return err
			}

			r.Header.Add("Authorization", adminBearerToken)

			resp, err = client.Do(r)
			if err != nil {
				return err
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			return nil
		})
	}
	err = errGroup.Wait()
	if err != nil {
		logger.PrintFatal(errors.New("unable to create accounts"), map[string]string{"error": err.Error()})
	}
	fmt.Printf("done creating accounts, took %v\n", time.Since(start))
	start = time.Now()

	// // create test transactions
	for i := 0; i < 200000; i++ {
		i := i
		errGroup.Go(func() error {
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

			resp, err := client.Do(r)
			if err != nil {
				logger.PrintFatal(errors.New("unable to perform new transfer request"), map[string]string{"error": err.Error()})
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()

			if i%1000 == 0 {
				fmt.Println(i)
			}
			return nil
		})
	}

	err = errGroup.Wait()
	if err != nil {
		logger.PrintFatal(errors.New("unable to create transfers"), map[string]string{"error": err.Error()})
	}
	fmt.Printf("done creating transfers, took %v\n", time.Since(start))
}
