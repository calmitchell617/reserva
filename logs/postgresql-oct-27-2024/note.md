# Testing PostgreSQL

Key points

- PostgreSQL running on Intel nuc 11 i7 with 3200 64GB DDR4 RAM
- Reserva running on Intel Nuc 11 i5 with 3200 64GB DDR4 RAM
- Concurrency of 32, running for X hours
- Single node setup
- Reserva and PostgreSQL servers connected via Netgear MS305 switch using 1 meter cat6a cables
    - Ping of around .65 millisecond
- Simulating "UK size economy"
    - 30 million account holders
    - 100 banks
    - prepopulating 100 million transactions

Config file as follows:

```
listen_addresses = '*'
shared_buffers = 16GB
effective_io_concurrency = 128
random_page_cost = 1.1
effective_cache_size = 32GB
```