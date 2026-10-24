-- 007: teams + team_seats for Bot team UUID roster / seat bind P0
CREATE TABLE IF NOT EXISTS teams (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  slug TEXT,
  template_key TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);
CREATE UNIQUE INDEX IF NOT EXISTS teams_ws_slug_uidx ON teams(workspace_id, slug) WHERE slug IS NOT NULL AND slug != '';

CREATE TABLE IF NOT EXISTS team_seats (
  id TEXT PRIMARY KEY,
  team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  role_slug TEXT NOT NULL,
  display_name TEXT NOT NULL,
  playbook_ref TEXT,
  bot_role_id TEXT REFERENCES bot_roles(id) ON DELETE SET NULL,
  bound_member_id TEXT REFERENCES members(id) ON DELETE SET NULL,
  bound_at TEXT,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(team_id, role_slug)
);
CREATE INDEX IF NOT EXISTS idx_team_seats_ws ON team_seats(workspace_id, team_id);
