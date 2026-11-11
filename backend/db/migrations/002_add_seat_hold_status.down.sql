BEGIN;
DROP INDEX IF EXISTS uq_active_seat_holds;
ALTER TABLE seat_holds
    DROP CONSTRAINT IF EXISTS chk_seat_holds_status;
ALTER TABLE seat_holds
    DROP COLUMN IF EXISTS released_at;
ALTER TABLE seat_holds
    DROP COLUMN IF EXISTS status;
COMMIT;

