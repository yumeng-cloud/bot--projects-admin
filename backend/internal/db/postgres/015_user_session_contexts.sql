-- 015: MOBILE_V2 MV2-G1 human session-context (one row per user; last-write-wins)
-- MOBILE_V2 MV2-G1: per-user session-context (human cross-device resume; not bot handoff)
CREATE TABLE IF NOT EXISTS user_session_contexts (
  user_id TEXT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  ws_id TEXT NOT NULL DEFAULT '',
  team_id TEXT NOT NULL DEFAULT '',
  project_id TEXT NOT NULL DEFAULT '',
  issue_id TEXT NOT NULL DEFAULT '',
  view TEXT NOT NULL DEFAULT '',
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);
