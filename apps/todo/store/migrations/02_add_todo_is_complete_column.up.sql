ALTER TABLE todos ADD COLUMN is_complete INTEGER NOT NULL DEFAULT 0;
UPDATE todos SET is_complete = 0 WHERE is_complete IS NULL;
