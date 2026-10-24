-- Postgres schema for AI collab board (pgvector for embeddings)
CREATE EXTENSION IF NOT EXISTS vector;


CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  name TEXT NOT NULL,
  avatar_url TEXT,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE TABLE IF NOT EXISTS workspaces (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE TABLE IF NOT EXISTS members (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL CHECK(role IN ('admin','member','bot')),
  display_name TEXT NOT NULL,
  is_bot INTEGER NOT NULL DEFAULT 0,
  bot_enabled INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, user_id)
);

CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  key_prefix TEXT NOT NULL,
  description TEXT DEFAULT '',
  archived_at TEXT,
  board_column_prefs TEXT,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, key_prefix)
);

CREATE TABLE IF NOT EXISTS issue_counters (
  project_id TEXT PRIMARY KEY REFERENCES projects(id) ON DELETE CASCADE,
  next_number INTEGER NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS labels (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  color TEXT NOT NULL DEFAULT '#1677ff',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, name)
);

CREATE TABLE IF NOT EXISTS cycles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  start_date TEXT,
  end_date TEXT,
  capacity INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'planned' CHECK(status IN ('planned','active','completed')),
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE TABLE IF NOT EXISTS issues (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  number INTEGER NOT NULL,
  identifier TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT DEFAULT '',
  status TEXT NOT NULL DEFAULT 'backlog'
    CHECK(status IN ('backlog','todo','in_progress','in_review','done','canceled')),
  priority TEXT NOT NULL DEFAULT 'none'
    CHECK(priority IN ('urgent','high','medium','low','none')),
  estimate INTEGER,
  assignee_id TEXT REFERENCES members(id) ON DELETE SET NULL,
  reporter_id TEXT REFERENCES members(id) ON DELETE SET NULL,
  cycle_id TEXT REFERENCES cycles(id) ON DELETE SET NULL,
  parent_id TEXT REFERENCES issues(id) ON DELETE SET NULL,
  -- P0 schedule: calendar day YYYY-MM-DD (Asia/Shanghai schedule day; not a TZ instant)
  start_at TEXT,
  due_at TEXT,
  board_order REAL NOT NULL DEFAULT 0,
  version INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(project_id, number),
  UNIQUE(workspace_id, identifier)
);

CREATE TABLE IF NOT EXISTS issue_labels (
  issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  label_id TEXT NOT NULL REFERENCES labels(id) ON DELETE CASCADE,
  PRIMARY KEY (issue_id, label_id)
);

CREATE TABLE IF NOT EXISTS issue_relations (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  from_issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  to_issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  relation_type TEXT NOT NULL CHECK(relation_type IN ('blocks','blocked_by','related','parent')),
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(from_issue_id, to_issue_id, relation_type)
);

CREATE TABLE IF NOT EXISTS comments (
  id TEXT PRIMARY KEY,
  issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  author_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  body TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE TABLE IF NOT EXISTS activities (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  issue_id TEXT REFERENCES issues(id) ON DELETE CASCADE,
  actor_id TEXT REFERENCES members(id) ON DELETE SET NULL,
  actor_kind TEXT NOT NULL DEFAULT 'human' CHECK(actor_kind IN ('human','bot','system')),
  action TEXT NOT NULL,
  payload TEXT DEFAULT '{}',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE TABLE IF NOT EXISTS api_keys (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  key_prefix TEXT NOT NULL,
  key_hash TEXT NOT NULL,
  last_used_at TEXT,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  revoked_at TEXT
);

CREATE TABLE IF NOT EXISTS webhooks (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  url TEXT NOT NULL,
  secret TEXT NOT NULL,
  events TEXT NOT NULL DEFAULT '["*"]',
  enabled INTEGER NOT NULL DEFAULT 1,
  filter_member_id TEXT,
  filter_seat_id TEXT,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
  id TEXT PRIMARY KEY,
  webhook_id TEXT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
  event TEXT NOT NULL,
  payload TEXT NOT NULL,
  status_code INTEGER,
  response_body TEXT,
  success INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE TABLE IF NOT EXISTS idempotency_keys (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL,
  key_value TEXT NOT NULL,
  method TEXT NOT NULL,
  path TEXT NOT NULL,
  response_status INTEGER NOT NULL,
  response_body TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, key_value)
);

CREATE INDEX IF NOT EXISTS idx_issues_project_status ON issues(project_id, status);
CREATE INDEX IF NOT EXISTS idx_issues_cycle ON issues(cycle_id);
CREATE INDEX IF NOT EXISTS idx_issues_assignee ON issues(assignee_id);
CREATE INDEX IF NOT EXISTS idx_issues_ws_start ON issues(workspace_id, start_at);
CREATE INDEX IF NOT EXISTS idx_issues_ws_due ON issues(workspace_id, due_at);
CREATE INDEX IF NOT EXISTS idx_activities_issue ON activities(issue_id, created_at);
CREATE INDEX IF NOT EXISTS idx_members_ws ON members(workspace_id);
CREATE INDEX IF NOT EXISTS idx_comments_issue ON comments(issue_id);

-- Wiki / Knowledge base
CREATE TABLE IF NOT EXISTS wiki_spaces (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  description TEXT DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, slug)
);

CREATE TABLE IF NOT EXISTS wiki_pages (
  id TEXT PRIMARY KEY,
  space_id TEXT NOT NULL REFERENCES wiki_spaces(id) ON DELETE CASCADE,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  parent_id TEXT REFERENCES wiki_pages(id) ON DELETE SET NULL,
  title TEXT NOT NULL,
  slug TEXT NOT NULL,
  body_md TEXT NOT NULL DEFAULT '',
  project_id TEXT REFERENCES projects(id) ON DELETE SET NULL,
  issue_id TEXT REFERENCES issues(id) ON DELETE SET NULL,
  created_by TEXT REFERENCES members(id) ON DELETE SET NULL,
  updated_by TEXT REFERENCES members(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(space_id, slug)
);

CREATE INDEX IF NOT EXISTS idx_wiki_pages_ws ON wiki_pages(workspace_id);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_title ON wiki_pages(workspace_id, title);

-- Time tracking
CREATE TABLE IF NOT EXISTS time_entries (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  hours REAL NOT NULL CHECK(hours > 0),
  note TEXT DEFAULT '',
  billable INTEGER NOT NULL DEFAULT 1,
  rate REAL,
  worked_at TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE INDEX IF NOT EXISTS idx_time_entries_issue ON time_entries(issue_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_member ON time_entries(member_id);
CREATE INDEX IF NOT EXISTS idx_time_entries_worked ON time_entries(workspace_id, worked_at);

CREATE TABLE IF NOT EXISTS billing_rates (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  member_id TEXT REFERENCES members(id) ON DELETE CASCADE,
  hourly_rate REAL NOT NULL DEFAULT 0,
  currency TEXT NOT NULL DEFAULT 'CNY',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, member_id)
);

-- Bot roles, memory, todos
CREATE TABLE IF NOT EXISTS bot_roles (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  slug TEXT NOT NULL,
  responsibilities TEXT DEFAULT '',
  skills TEXT DEFAULT '[]',
  default_project_ids TEXT DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, slug)
);

CREATE TABLE IF NOT EXISTS bot_role_claims (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  role_id TEXT NOT NULL REFERENCES bot_roles(id) ON DELETE CASCADE,
  member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  claimed_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  released_at TEXT,
  UNIQUE(role_id) -- one active claim: enforced in app by released_at IS NULL
);

CREATE TABLE IF NOT EXISTS bot_memories (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  role_id TEXT REFERENCES bot_roles(id) ON DELETE SET NULL,
  key_name TEXT NOT NULL,
  value_text TEXT NOT NULL,
  tags TEXT DEFAULT '[]',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(workspace_id, member_id, key_name)
);

CREATE INDEX IF NOT EXISTS idx_bot_memories_search ON bot_memories(workspace_id, member_id);

CREATE TABLE IF NOT EXISTS bot_todos (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  member_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  role_id TEXT REFERENCES bot_roles(id) ON DELETE SET NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'todo' CHECK(status IN ('todo','doing','done','canceled')),
  issue_id TEXT REFERENCES issues(id) ON DELETE SET NULL,
  sort_order REAL NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE INDEX IF NOT EXISTS idx_bot_todos_member ON bot_todos(member_id, status);

-- api_key hash uniqueness (dual-driver)
CREATE UNIQUE INDEX IF NOT EXISTS api_keys_key_hash_uidx ON api_keys(key_hash);


-- Attachments (issue / comment files on disk; metadata only in DB)
CREATE TABLE IF NOT EXISTS attachments (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  issue_id TEXT NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
  comment_id TEXT REFERENCES comments(id) ON DELETE CASCADE,
  uploader_id TEXT NOT NULL REFERENCES members(id) ON DELETE CASCADE,
  filename TEXT NOT NULL,
  content_type TEXT NOT NULL,
  size_bytes INTEGER NOT NULL,
  storage_key TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'))
);

CREATE INDEX IF NOT EXISTS idx_attachments_issue ON attachments(workspace_id, issue_id);
CREATE INDEX IF NOT EXISTS idx_attachments_comment ON attachments(workspace_id, comment_id);


-- Teams + seats (Bot team UUID roster / seat bind P0)
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
  status TEXT NOT NULL DEFAULT 'open',
  created_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  updated_at TEXT NOT NULL DEFAULT (to_char(now() AT TIME ZONE 'utc', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')),
  UNIQUE(team_id, role_slug)
);
CREATE INDEX IF NOT EXISTS idx_team_seats_ws ON team_seats(workspace_id, team_id);

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

-- Team pins (shared memory wall per team; isolated from bot_memories)
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


-- Shared embedding chunks (Wiki + Bot memory); dims must match EMBED_DIMS (default 384)
CREATE TABLE IF NOT EXISTS embedding_chunks (
  id TEXT PRIMARY KEY,
  workspace_id TEXT NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
  entity_type TEXT NOT NULL,
  entity_id TEXT NOT NULL,
  chunk_idx INT NOT NULL DEFAULT 0,
  content TEXT NOT NULL DEFAULT '',
  embedding vector(384) NOT NULL,
  model_id TEXT NOT NULL,
  dims INT NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (workspace_id, entity_type, entity_id, chunk_idx)
);
CREATE INDEX IF NOT EXISTS idx_embedding_chunks_entity ON embedding_chunks (workspace_id, entity_type, entity_id);
CREATE INDEX IF NOT EXISTS idx_embedding_chunks_hnsw ON embedding_chunks USING hnsw (embedding vector_cosine_ops);

-- Keyword assist: trigram optional; hybrid uses ILIKE + vector RRF in app
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_wiki_pages_title_trgm ON wiki_pages USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_wiki_pages_body_trgm ON wiki_pages USING gin (body_md gin_trgm_ops);
