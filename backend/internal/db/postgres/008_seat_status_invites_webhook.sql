-- 008: team_seats.status + team_invites (ymt_) + webhook seat/member filters (G2/G3/G6a)
ALTER TABLE team_seats ADD COLUMN IF NOT EXISTS status TEXT NOT NULL DEFAULT 'open';
UPDATE team_seats SET status = 'active' WHERE bound_member_id IS NOT NULL AND bound_member_id != '' AND status = 'open';

CREATE TABLE IF NOT EXISTS team_invites (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  team_id TEXT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
  token_hash TEXT NOT NULL UNIQUE,
  token_prefix TEXT NOT NULL DEFAULT '',
  allow_expand_seat INTEGER NOT NULL DEFAULT 0,
  seat_slug TEXT,
  role_hint TEXT,
  max_uses INTEGER,
  use_count INTEGER NOT NULL DEFAULT 0,
  expires_at TEXT,
  revoked_at TEXT,
  created_by TEXT REFERENCES members(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);
CREATE INDEX IF NOT EXISTS idx_team_invites_ws_team ON team_invites(workspace_id, team_id);
CREATE INDEX IF NOT EXISTS idx_team_invites_hash ON team_invites(token_hash);

ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS filter_member_id TEXT;
ALTER TABLE webhooks ADD COLUMN IF NOT EXISTS filter_seat_id TEXT;
