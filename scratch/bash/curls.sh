# Get an auth token
curl -d '{"username": "adminBank", "password": "Mypass123"}' http://localhost/v1/tokens/authentication
curl -d '{"username": "bulgarubank", "password": "Mypass123"}' localhost/v1/tokens/authentication
FBQOSPPLCWPXVQEX2FULXUUGOI

# Create a bank
curl -H 'Authorization: Bearer 6DNWEM4E6ZFLJPH2I73YHBRBYE' -d '{"username": "bulgarubank", "admin": false, "password": "Mypass123"}' localhost/v1/banks

# Get a bank
curl -H 'Authorization: Bearer 6DNWEM4E6ZFLJPH2I73YHBRBYE' "localhost/v1/banks/bulgarubank"

# Create an account
curl -H 'Authorization: Bearer 6DNWEM4E6ZFLJPH2I73YHBRBYE' -d '{"metadata": "{\"1\": 4}"}' localhost/v1/accounts

# Get an account
curl -H 'Authorization: Bearer 6DNWEM4E6ZFLJPH2I73YHBRBYE' localhost/v1/accounts/1

# Freeze an account
curl -X PATCH -d '{"id": 1, "frozen": false}' -H 'Authorization: Bearer 6DNWEM4E6ZFLJPH2I73YHBRBYE' localhost/v1/accounts/frozen

# Change an account's balance
curl -H 'Authorization: Bearer 6DNWEM4E6ZFLJPH2I73YHBRBYE' -X PATCH -d '{"id": 1, "change_in_cents": 450}' localhost/v1/accounts/change_money_supply

# Create a transfer 
curl -H 'Authorization: Bearer M74NMJ2TIHGSEZ3VGIWNNVKDNE' -d '{"source_account_id": 1, "target_account_id": 2, "amount_in_cents": 100}' http://localhost/v1/transfers