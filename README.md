# Reserva Readme
Reserva is a functional central bank digital currency (CBDC) running on top of the SingleStore database. It is written in the Go programming language and exposes its functionality via a REST API.

Reserva is a stateless web server, and thus is horizontally scalable, and suitable for deployment in containers. 

You can read more about Reserva at [my website's blog](https://www.sqlpipe.com/blog/reserva-cbdc-singlestore).

# API docs

All routes require bearer token authentication unless otherwise noted.

Any bank that is given "admin" permissions has the ability to change the money supply, which makes it a "central bank" in traditional economic terms.

## Banks
---

### `/v1/banks`
- `POST`
  - Creates a new bank
  - Requires admin token

  ### ***Request***
  ```
  {
    "username": <string>,
    "admin": <boolean>
  }
  ```

  ### ***Response***
  ```
  {
    "username": <string>,
    "admin": <boolean>
  }
  ```

### `/v1/banks/:username`
- `GET`
  - Gets the bank's info
  - Requires bank's token, or admin token

  ### ***Request***
  `GET` with no query parameters.

  ### ***Response***
  ```
  {
    "username": <string>,
    "admin": <boolean>
  }
  ```

## Accounts

- `POST`
  - Creates an account

  ### ***Request***
  ```
  {
    "metadata": <object>
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "metadata": <object>,
    "bank_username": <string>,
    "frozen": <boolean>,
    "balance_in_cents": <number>
  }
  ```

### `/v1/accounts/:id`
- `GET`
  - Gets an account's info

  ### ***Request***
  `GET` with no query parameters.

  ### ***Response***
  ```
  {
    "id": <number>,
    "metadata": <object>,
    "bank_username": <string>,
    "frozen": <boolean>,
    "balance_in_cents": <number>
  }
  ```

### `/v1/accounts/balance_in_cents`
- `PATCH`
  - Changes the account's balance, which alters the total money supply.
  - Requires admin token

  ### ***Request***
  ```
  {
    "id": <number>,
    "amount_in_cents": <number>
  }

  ```
  ### ***Response***
  ```
  {
    "id": <number>,
    "metadata": <object>,
    "bank_username": <string>,
    "frozen": <boolean>,
    "balance_in_cents": <number>
  }
  ```

### `/v1/accounts/metadata`
- `PATCH`
  - Changes account's KYC data.

  ### ***Request***
  ```
  {
    "id": <number>,
    "metadata": <object>
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "metadata": <object>,
    "bank_username": <string>,
    "frozen": <boolean>,
    "balance_in_cents": <number>
  }
  ```

### `/v1/accounts/frozen`
- `PATCH`
  - Freezes or unfreezes an account.

  ### ***Request***
  ```
  {
    "id": <number>,
    "frozen": <boolean>
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "metadata": <object>,
    "bank_username": <string>,
    "frozen": <boolean>,
    "balance_in_cents": <number>
  }
  ```

## Transfers

### `/v1/transfers`
- `POST`
  - Create a new transfer.

  ### ***Request***
  ```
  {
    "source_account_id": <number>,
    "target_account_id": <number>,
    "amount_in_cents": <number>
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "created_at": <string...RFC 3339>
    "source_account": <number>,
    "target_account": <number>,
    "amount_in_cents": <number>
  }
  ```

## Tokens

### `/v1/tokens/authentication`
- `POST`
  - Create an authentication token

  ### ***Request***
  ```
  {
    "email": <string>,
    "password": <string>
  }
  ```

  ### ***Response***
  ```
  {
    "token": <string>
  }
  ```

## Utility

### `/v1/healthcheck`
- `GET`
  - A healthcheck

### `/debug/vars`
- `GET`
  - Get system info