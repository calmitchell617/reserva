# Testing PostgreSQL

Key points

- PostgreSQL running on Intel nuc 11 i7 with 3200 64GB DDR4 RAM
- Reserva running on Intel Nuc 11 i5 with 3200 64GB DDR4 RAM
- Concurrency of 64, running for 1 hour
- Single node setup
- deletes enabled
- not using kinda random
- Reserva and PostgreSQL servers connected via Netgear MS305 switch using 1 meter cat6a cables
    - Ping of around .65 millisecond
- Simulating "Portugal size economy"
    - 50 banks
    - 10 million account holders
    - 10 million pre existing transactions

Config file as follows:

```
listen_addresses = '*'

shared_buffers = 16GB
```

