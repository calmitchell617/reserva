# Get an auth token
curl -d '{"username": "calBank", "password": "Mypass123"}' localhost/v1/tokens/authentication
curl -d '{"username": "bulgarubank", "password": "Mypass123"}' localhost/v1/tokens/authentication

# Create a bank
curl -H 'Authorization: Bearer 62UEP4WJUW3WWO6ZSC42WAALA4' -d '{"username": "bulgarubank", "admin": false, "password": "Mypass123"}' localhost/v1/banks

# Get a bank
curl -H 'Authorization: Bearer I7PPRAYM3MZK2BPDU7NLRF7BF4' "localhost/v1/banks?username=calBank"