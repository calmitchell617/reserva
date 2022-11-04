# Get an auth token
curl -d '{"username": "adminBank", "password": "Mypass123"}' http://172.31.72.165/v1/tokens/authentication
curl -d '{"username": "bulgarubank", "password": "Mypass123"}' localhost/v1/tokens/authentication

# Create a bank
curl -H 'Authorization: Bearer QQHUKGIXTW4D25G3INWJ7WN3WM' -d '{"username": "bulgarubank", "admin": false, "password": "Mypass123"}' localhost/v1/banks

# Get a bank
curl -H 'Authorization: Bearer MIVSBTCE3KFOTIST34KRBXDKJE' "localhost/v1/banks/bulgarubank"

# Create an account
curl -H 'Authorization: Bearer MIVSBTCE3KFOTIST34KRBXDKJE' -d '{"metadata": "{\"1\": 4}"}' localhost/v1/accounts

# Get an account
curl -H 'Authorization: Bearer MIVSBTCE3KFOTIST34KRBXDKJE' localhost/v1/accounts/1

# Freeze an account
curl -X PATCH -d '{"id": 1, "frozen": false}' -H 'Authorization: Bearer MIVSBTCE3KFOTIST34KRBXDKJE' localhost/v1/accounts/frozen

# Change an account's balance
curl -H 'Authorization: Bearer QQHUKGIXTW4D25G3INWJ7WN3WM' -X PATCH -d '{"id": 1, "change_in_cents": 450}' localhost/v1/accounts/change_money_supply

# Create a transfer 
curl -H 'Authorization: Bearer G3UOWGQVIFT77GEZJS47PLODZE' -d '{"source_account_id": 1, "target_account_id": 2, "amount_in_cents": 100}' http://172.31.72.165/v1/transfers