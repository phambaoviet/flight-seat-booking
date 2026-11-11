BEGIN;
ALTER TABLE seat_holds
    ADD COLUMN status TEXT NOT NULL DEFAULT 'ACTIVE';
ALTER TABLE seat_holds
    ADD COLUMN released_at TIMESTAMPTZ;
UPDATE seat_holds
SET status = 'EXPIRED'
WHERE expires_at <= CURRENT_TIMESTAMP;
ALTER TABLE seat_holds
    ADD CONSTRAINT chk_seat_holds_status
    CHECK (status IN ('ACTIVE', 'EXPIRED', 'RELEASED', 'CONVERTED'));
CREATE UNIQUE INDEX uq_active_seat_holds
    ON seat_holds (seat_id)
    WHERE status = 'ACTIVE';
COMMIT;
