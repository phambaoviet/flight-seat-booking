CREATE TABLE users (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone_number TEXT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE flights(
    id UUID PRIMARY KEY,
    flight_number TEXT NOT NULL,
    departure_airport TEXT NOT NULL,
    arrival_airport TEXT NOT NULL,
    boarding_at TIMESTAMPTZ NOT NULL,
    departure_at TIMESTAMPTZ NOT NULL,
    arrival_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CHECK (boarding_at < departure_at),
    CHECK (departure_at < arrival_at),
    CHECK (departure_airport <> arrival_airport),
    CHECK (departure_airport ~ '^[A-Z]{3}$'),
    CHECK (arrival_airport ~ '^[A-Z]{3}$')

);
CREATE TABLE seats(
    id UUID PRIMARY KEY,
    flight_id UUID REFERENCES flights(id) ON DELETE CASCADE,
    seat_number TEXT NOT NULL,
    seat_class TEXT NOT NULL,
    price BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    UNIQUE(flight_id, seat_number),

    CHECK (price >= 0)
);
CREATE TABLE seat_holds(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    seat_id UUID NOT NULL REFERENCES seats(id) ON DELETE CASCADE,
    hold_token UUID UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ NOT NULL,

    CHECK (expires_at > created_at)
);
CREATE TABLE bookings(
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id),
    seat_id UUID NOT NULL REFERENCES seats(id),
    booking_code TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CHECK (status IN ('CONFIRMED', 'CANCELLED'))
);

CREATE INDEX idx_flights_search 
ON flights(
    departure_airport, 
    arrival_airport, 
    departure_at
);

CREATE INDEX idx_seat_holds_seats_expiry
ON seat_holds(
    seat_id,
    expires_at
);

CREATE UNIQUE INDEX idx_one_confirmed_booking_per_seat
ON bookings(
    seat_id
)
WHERE status = 'CONFIRMED';

