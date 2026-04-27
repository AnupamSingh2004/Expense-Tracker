CREATE INDEX IF NOT EXISTS idx_expenses_category      ON expenses (category);
CREATE INDEX IF NOT EXISTS idx_expenses_date          ON expenses (date DESC);
CREATE INDEX IF NOT EXISTS idx_expenses_category_date ON expenses (category, date DESC);
