# Reserva

Reserva is a database benchmarking tool that simulates a streamlined, high volume payments processing system. It is written in Go and supports arbitrary amounts of concurrency.

On start, Reserva loads all users and authentication tokens into memory, then attempts to make as many funds transfers as possible in a loop. Each funds transfer requires 4 round trips to the database and touches 6 tables. The workflow is as follows:

1. DB is queried to authenticate and authorize the user who is requesting payment.
2. DB is queried to get information about the card used and account to be debited.
3. DB is queried to authenticate and authorize the user who will approve or deny the payment request.
4. DB is queried to update both account balances and record the payment.

Reserva allows you to select whether or not you want deletes to be part of the workload.
