DROP INDEX IF EXISTS idx_expenses_user_id;
ALTER TABLE expenses DROP COLUMN IF EXISTS user_id;
