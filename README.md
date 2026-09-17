# Flight Seat Booking Engine

A high-concurrency seat reservation and booking backend service built with Go, PostgreSQL, and SQLC.

The system focuses on preventing double-booking under concurrent requests, enforcing ACID transactions during seat reservation, and handling temporary seat holds using lazy expiration.

---

## Key Technical Problems Solved

### Concurrency & Double-Booking Prevention

Prevents race conditions when multiple users attempt to reserve the same seat simultaneously.

Implemented using:

- PostgreSQL row-level pessimistic locking
- `SELECT ... FOR UPDATE`
- Atomic database transactions
- Database constraints for confirmed bookings

### Seat Reservation State Machine

The system supports the following logical state transitions:

```text
                 ┌──────────────────────────────┐
                 │ Hold Expired / Cancelled     │
                 └──────────────┬───────────────┘
                                │
                                ▼
[ AVAILABLE ] ──(Hold / 10 min)──► [ HOLDING ]
                                      │
                                      │ Payment Success
                                      ▼
                                  [ BOOKED ]

[ HOLDING ] ──(Expired / Cancelled)──► [ AVAILABLE ]
```
