package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/calmitchell617/reserva/internal/jsonlog"
	"golang.org/x/sync/errgroup"
)

var (
	host         string
	concurrency  int
	numTransfers int
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

	// where is the API server?
	flag.StringVar(&host, "host", "", "reserva host")
	// How many concurrent requests to the API server do you want to make?
	flag.IntVar(&concurrency, "concurrency", 100, "max num requests")
	flag.IntVar(&numTransfers, "num-transfers", 100000, "number of test transfers to run")

	flag.Parse()

	if host == "" {
		fmt.Println("You must enter a host to load test")
		os.Exit(1)
	}

	// establish json logging
	logger := jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// Create a client that has a maximum number of concurrent open connections
	// if you don't do this, you will run out of ports
	client := &http.Client{Transport: &http.Transport{
		MaxIdleConnsPerHost: 20,
		MaxConnsPerHost:     concurrency * 2,
	}}

	// Get auth bearer token for admin bank
	json_data, err := json.Marshal(map[string]string{"username": "adminBank", "password": "Mypass123"})
	if err != nil {
		logger.PrintFatal(errors.New("unable to marshal json data"), map[string]string{"error": err.Error()})
	}

	resp, err := client.Post(fmt.Sprintf("%v/v1/tokens/authentication", host), "application/json; charset=UTF-8", bytes.NewBuffer(json_data))
	if err != nil {
		logger.PrintFatal(errors.New("unable to get initial admin bank auth token"), map[string]string{"error": err.Error()})
	}

	if resp.StatusCode != http.StatusCreated {
		logger.PrintFatal(errors.New("unable to get initial admin bank auth token"), map[string]string{})
	}

	authTokenResp := AuthTokenResponse{}
	json.NewDecoder(resp.Body).Decode(&authTokenResp)
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()
	adminBearerToken := fmt.Sprintf("Bearer %v", authTokenResp.AuthToken.Token)
	fmt.Printf("finished setting up admin bank, took %v\n", time.Since(start))

	// Create some test banks
	numBanks := 100
	bankBearerTokens := make([]string, numBanks)

	// create an errGroup to do concurrent request processing, while handling errors
	errGroup := new(errgroup.Group)
	errGroup.SetLimit(concurrency)

	start = time.Now()
	for i := 0; i < numBanks; i++ {
		// This looks silly, but if you don't do i := i, then
		// the goroutine won't always use the correct value of i
		i := i
		errGroup.Go(func() error {
			// create a bank
			json_data, err := json.Marshal(map[string]any{"username": fmt.Sprintf("bank%v", i), "admin": false, "password": "Mypass123"})
			if err != nil {
				return err
			}
			r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/banks", host), bytes.NewBuffer(json_data))
			if err != nil {
				return err
			}

			r.Header.Add("Authorization", adminBearerToken)

			// do the request
			resp, err := client.Do(r)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusAccepted {
				logger.PrintFatal(errors.New("unable to create bank"), map[string]string{})
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

			if resp.StatusCode != http.StatusAccepted {
				logger.PrintFatal(errors.New("unable to get admin auth token"), map[string]string{})
			}

			authTokenResp := AuthTokenResponse{}
			json.NewDecoder(response.Body).Decode(&authTokenResp)
			io.Copy(ioutil.Discard, resp.Body)
			response.Body.Close()
			bearerToken := fmt.Sprintf("Bearer %v", authTokenResp.AuthToken.Token)
			// store it in our list of bearer tokens to be used later
			bankBearerTokens[i] = bearerToken
			return nil
		})
	}
	// check to make sure no errors happened
	err = errGroup.Wait()
	if err != nil {
		logger.PrintFatal(errors.New("unable to create banks"), map[string]string{"error": err.Error()})
	}
	fmt.Printf("done creating %v banks, took %v\n", numBanks, time.Since(start))
	start = time.Now()

	// create test accounts
	numAccounts := 1000
	for i := 0; i < 1000; i++ {
		i := i
		errGroup.Go(func() error {
			// create account data, with a random lat/long
			json_data, err := json.Marshal(map[string]any{"metadata": fmt.Sprintf(`{"lat": %v, "lng": %v}`, rand.Intn(45), rand.Intn(45))})
			if err != nil {
				return err
			}
			r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/accounts", host), bytes.NewBuffer(json_data))
			if err != nil {
				return err
			}

			// get a random bearer token, and therefore give the account a random controlling bank
			randomBearerTokenIndex := int(math.Sqrt(float64(rand.Intn(numBanks))))
			r.Header.Add("Authorization", bankBearerTokens[randomBearerTokenIndex])

			resp, err := client.Do(r)
			if err != nil {
				return err
			}
			if resp.StatusCode != http.StatusCreated {
				logger.PrintFatal(errors.New("unable to get auth token for bank"), map[string]string{})
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()

			time.Sleep(time.Second * 1)

			// change money supply by adding to account
			data := map[string]any{"id": i + 1, "change_in_cents": 100000000}
			json_data, err = json.Marshal(data)
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
			if resp.StatusCode != http.StatusOK {

				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)
				fmt.Println(data)
				logger.PrintFatal(errors.New("unable to change money supply"), map[string]string{"error": bodyString})
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
			return nil
		})
	}
	// check for errors while creating accounts
	err = errGroup.Wait()
	if err != nil {
		logger.PrintFatal(errors.New("unable to create accounts"), map[string]string{"error": err.Error()})
	}
	fmt.Printf("done creating %v accounts, took %v\n", numAccounts, time.Since(start))
	start = time.Now()

	// create test transactions
	for i := 0; i < numTransfers; i++ {
		i := i
		errGroup.Go(func() error {
			// get a random source and target account id
			min := 1
			max := 1000
			sourceIdSeed := rand.Intn(max-min) + min
			targetIdSeed := rand.Intn(max-min) + min
			sourceId := int(math.Sqrt(float64(sourceIdSeed)))
			targetId := int(math.Sqrt(float64(targetIdSeed)))

			// get a random amount of money to transfer
			min = 10
			max = 30
			amountInCents := rand.Intn(max-min) + min

			data := map[string]any{"source_account_id": sourceId, "target_account_id": targetId, "amount_in_cents": amountInCents}
			json_data, err := json.Marshal(data)
			if err != nil {
				logger.PrintFatal(errors.New("unable to marshal new transfer json data"), map[string]string{"error": err.Error()})
			}
			r, err := http.NewRequest("POST", fmt.Sprintf("%v/v1/transfers", host), bytes.NewBuffer(json_data))
			if err != nil {
				logger.PrintFatal(errors.New("unable to create new transfer request"), map[string]string{"error": err.Error()})
			}

			r.Header.Add("Authorization", adminBearerToken)

			// make the request
			resp, err := client.Do(r)
			if err != nil {
				logger.PrintFatal(errors.New("unable to perform new transfer request"), map[string]string{"error": err.Error()})
			}
			if resp.StatusCode != http.StatusCreated {
				bodyBytes, err := io.ReadAll(resp.Body)
				if err != nil {
					log.Fatal(err)
				}
				bodyString := string(bodyBytes)
				fmt.Println(data)
				logger.PrintFatal(errors.New("unable to make transfer"), map[string]string{"error": bodyString})
			}
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()

			// every 1000 transfers, print a status update
			if i%1000 == 0 {
				fmt.Printf("finished %v transfers\n", i)
			}
			return nil
		})
	}

	// check transfer errors
	err = errGroup.Wait()
	if err != nil {
		logger.PrintFatal(errors.New("unable to create transfers"), map[string]string{"error": err.Error()})
	}
	fmt.Printf("done creating %v transfers, took %v, at a rate of %v per second\n", numTransfers, time.Since(start), numTransfers/int(time.Duration(time.Since(start)).Seconds()))
}
