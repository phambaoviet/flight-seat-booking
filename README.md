# Flight Seat Booking Engine

A high-concurrency seat reservation and booking backend built with Go, PostgreSQL, and SQLC.

The system focuses on preventing double-booking under concurrent requests, enforcing ACID transactions during seat reservation, and handling temporary seat holds through lazy expiration.

## Key Technical Problems Solved

### 1. Concurrency and Double-Booking Prevention

Prevents race conditions when multiple users attempt to reserve the same seat simultaneously.

Implemented using:

- PostgreSQL row-level pessimistic locking
- `SELECT ... FOR UPDATE`
- Atomic database transactions
- Database constraints for confirmed bookings
- Concurrent integration tests

### 2. Seat Reservation State Machine

The system supports the following logical state transitions:

```text
                  Hold Expired / Cancelled
                           |
                           v
[ AVAILABLE ] -- Hold / 10 min --> [ HOLDING ]
                                        |
                                        | Confirmation
                                        v
                                    [ BOOKED ]

[ HOLDING ] -- Expired / Cancelled --> [ AVAILABLE ]
```

Seat availability is determined from active holds and confirmed bookings rather than relying on a permanently stored seat status.

### 3. Transactional Booking Confirmation

Booking confirmation is handled within a database transaction.

The transaction:

1. Locks the target seat.
2. Validates the user's active seat hold.
3. Creates a confirmed booking.
4. Deletes the active hold.
5. Commits the transaction only if all operations succeed.

If an operation fails, the transaction is rolled back.

### 4. Lazy Expiration

Seat holds expire after 10 minutes.

Expired holds are treated as inactive by database queries, so the system does not depend on a heavy periodic cron job to determine seat availability.

### 5. Database Design and Query Optimization

- PostgreSQL as the source of truth
- SQL migrations for schema management
- SQLC for type-safe Go database access
- Composite indexes for flight search and seat-hold queries
- A partial unique index to prevent multiple confirmed bookings for the same seat

## Tech Stack

- **Language:** Go
- **Database:** PostgreSQL
- **Database Access:** SQLC, pgx
- **Database Migrations:** golang-migrate
- **Testing:** Go testing, Testify
- **Containerization:** Docker, Docker Compose

## Project Structure

```text
flight-seat-booking/
├── db/
│   ├── migrations/        # Database schema migrations
│   ├── query/             # SQL queries used by SQLC
│   └── sqlc/              # SQLC-generated Go code
├── internal/
│   ├── store/             # Database operations and transactions
│   ├── service/           # Business logic
│   └── handler/            # HTTP request handlers
├── docker-compose.yml
├── Makefile
├── go.mod
└── README.md
```

## Testing

Run all Go tests:

```bash
go test ./...
```

The test suite includes integration tests for seat holds, booking confirmation, transaction errors, and concurrent attempts to hold the same seat.

## Current Development Focus

- Separating generated database code from custom Store logic
- Building the Service Layer with dependency injection
- Testing business logic independently from PostgreSQL
- Integrating HTTP handlers with the Service Layer

## Design Principles

- PostgreSQL is the source of truth for seat availability.
- Critical reservation operations must be transaction-safe.
- Expired holds must not block seats indefinitely.
- Database constraints provide an additional layer of correctness.
- Generated SQLC code is kept separate from custom application logic.
