-- 006: S4 project board column prefs as JSON column (no separate table)
ALTER TABLE projects ADD COLUMN IF NOT EXISTS board_column_prefs TEXT;
