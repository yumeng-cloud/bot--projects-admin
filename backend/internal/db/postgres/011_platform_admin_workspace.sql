-- 011: platform admin flag, workspace disable, human workspace invites, audit
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_platform_admin INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workspaces ADD COLUMN IF NOT EXISTS disabled_at TEXT;

CREATE TABLE IF NOT EXISTS workspace_invites (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  email TEXT,
  role TEXT NOT NULL CHECK(role IN ('admin','member')),
  token_hash TEXT NOT NULL UNIQUE,
  token_prefix TEXT NOT NULL DEFAULT '',
  max_uses INTEGER,
  use_count INTEGER NOT NULL DEFAULT 0,
  expires_at TEXT,
  revoked_at TEXT,
  created_by TEXT REFERENCES members(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);
CREATE INDEX IF NOT EXISTS idx_workspace_invites_ws ON workspace_invites(workspace_id);
CREATE INDEX IF NOT EXISTS idx_workspace_invites_hash ON workspace_invites(token_hash);

CREATE TABLE IF NOT EXISTS platform_audit_logs (
  id TEXT PRIMARY KEY,
  actor_user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  action TEXT NOT NULL,
  target_type TEXT NOT NULL,
  target_id TEXT NOT NULL,
  payload TEXT DEFAULT '{}',
  created_at TEXT NOT NULL DEFAULT (to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);
CREATE INDEX IF NOT EXISTS idx_platform_audit_created ON platform_audit_logs(created_at);
