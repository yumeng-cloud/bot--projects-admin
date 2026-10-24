-- 010: Issue pool by team (T6-I). Nullable team_id; NULL = workspace-wide visible (legacy).
ALTER TABLE issues ADD COLUMN IF NOT EXISTS team_id TEXT REFERENCES teams(id) ON DELETE SET NULL;
CREATE INDEX IF NOT EXISTS idx_issues_ws_team ON issues(workspace_id, team_id);
