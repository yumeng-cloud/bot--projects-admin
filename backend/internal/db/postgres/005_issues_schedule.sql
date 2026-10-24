-- 005: Issue schedule fields for Timeline P0 (calendar day YYYY-MM-DD)
ALTER TABLE issues ADD COLUMN IF NOT EXISTS start_at TEXT;
ALTER TABLE issues ADD COLUMN IF NOT EXISTS due_at TEXT;
CREATE INDEX IF NOT EXISTS idx_issues_ws_start ON issues(workspace_id, start_at);
CREATE INDEX IF NOT EXISTS idx_issues_ws_due ON issues(workspace_id, due_at);
