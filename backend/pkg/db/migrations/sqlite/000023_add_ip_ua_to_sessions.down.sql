-- SQLite does not support DROP COLUMN before 3.35.0.
-- The columns are intentionally left in place on rollback.
SELECT 1;
