-- 012: Issue archive (ISSUE_ARCHIVE). Nullable archived_at; NULL = active (default list).
ALTER TABLE issues ADD COLUMN IF NOT EXISTS archived_at TEXT;
CREATE INDEX IF NOT EXISTS idx_issues_ws_archived ON issues(workspace_id, archived_at);
