-- 009: team_pins (shared memory wall per team; isolated from bot_memories)
CREATE TABLE IF NOT EXISTS team_pins (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  key TEXT NOT NULL,
  value TEXT NOT NULL DEFAULT '',
  updated_by TEXT REFERENCES members(id) ON DELETE SET NULL,
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(team_id, key)
);
CREATE INDEX IF NOT EXISTS idx_team_pins_ws_team ON team_pins(workspace_id, team_id);
