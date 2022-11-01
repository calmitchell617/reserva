# Reserva public API

All routes require bearer token authentication unless otherwise noted.

## Banks
---

### `/v1/banks`
- `GET`
  - Gets the requesting bank's info.
  - Requires bank's token, or admin token

  ### ***Request***
  `GET` with `?id=<number>` query parameter.

  ### ***Response***
  ```
  {
    "id": <number>,
    "username": <string>,
    "balance_in_cents": <number>,
    "frozen": <boolean>
  }
  ```

- `POST`
  - Creates a new bank
  - Requires admin token

  ### ***Request***
  ```
  {
    "username": <string>,
    "name": <string>
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "username": <string>,
    "name": <string>,
    "balance_in_cents": <number>,
    "frozen": <boolean>
  }
  ```

### `/v1/banks/balance_in_cents`
- `PATCH`
  - Updates bank's balance
  - Requires admin token

  ### ***Request***
  ```
  {
    "id": <number>,
    "balance_in_cents": <number>
  }

  ```
  ### ***Response***
  ```
  {
    "id": <number>,
    "username": <string>,
    "name": <string>,
    "balance_in_cents": <number>,
    "frozen": <boolean>
  }
  ```

### `/v1/banks/frozen`
- `PATCH`
  - Updates bank's frozen value
  - Requires admin token

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
    "username": <string>,
    "name": <string>,
    "balance_in_cents": <number>,
    "frozen": <boolean>
  }
  ```

## Accounts

### `/v1/accounts`
- `GET`
  - Gets an account's info

  ### ***Request***
  `GET` with `?id=<number>` query parameter.

  ### ***Response***
  ```
  {
    "id": <number>,
    "kyc_data": <object>,
    "bank_id": <number>,
    "frozen": <boolean>,
    "balance_in_cents": <number>
  }
  ```

- `POST`
  - Creates an account

  ### ***Request***
  ```
  {
    "kyc_data": <object>
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "kyc_data": <object>,
    "bank_id": <number>,
    "frozen": <boolean>,
    "balance_in_cents": <number>
  }
  ```

### `/v1/accounts/kyc_data`
- `PATCH`
  - Changes account's KYC data.

  ### ***Request***
  ```
  {
    "id": <number>,
    "kyc_data": <object>
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "kyc_data": <object>,
    "bank_id": <number>,
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
    "kyc_data": <object>,
    "bank_id": <number>,
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
    "amount_in_cents": <number>,
    "latitude": <number>,
    "longitude": <number> 
  }
  ```

  ### ***Response***
  ```
  {
    "id": <number>,
    "created_at": <string...RFC 3339>
    "source_account": <number>,
    "target_account": <number>,
    "amount_in_cents": <number>,
    "latitude": <number>,
    "longitude": <number>
  }
  ```

- `GET`
  - Allows you to search for transactions with various filtering critera.

  ### ***Request***
  `GET` with `id`, `source_account`, or `target_account` query parameters. Also can filter with `created_at`.

  ### ***Response***
  ```
  [
    {
      "id": <number>,
      "created_at": <string...RFC 3339>
      "source_account": <number>,
      "target_account": <number>,
      "amount_in_cents": <number>
    },
    {
      "id": <number>,
      "created_at": <string...RFC 3339>
      "source_account": <number>,
      "target_account": <number>,
      "amount_in_cents": <number>
    }...
  ]
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

### `/v1/tokens/activation`
- `POST`
  - Create an activation token

  ### ***Request***
  ```
  {
    "email": <string>
  }
  ```

  ### ***Response***
  ```
  {
    "message": <string>
  }
  ```

### `/v1/tokens/password-reset`
- `POST`
  - Create a password reset token

  ### ***Request***
  ```
  {
    "email": <string>
  }
  ```

  ### ***Response***
  ```
  {
    "message": <string>
  }
  ```

## Utility

### `/v1/healthcheck`
- `GET`
  - A healthcheck

  ### ***Request***
  `GET` with no query parameters.

  ### ***Response***
  ```
  {
		"status": <string>,
		"system_info": {
			"environment": <string>,
			"version": <string
		}
	}
  ```