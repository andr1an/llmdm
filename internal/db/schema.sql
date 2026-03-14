-- campaigns
CREATE TABLE IF NOT EXISTS campaigns (
  id              TEXT PRIMARY KEY,
  name            TEXT NOT NULL,
  description     TEXT,
  created_at      TEXT DEFAULT CURRENT_TIMESTAMP,
  current_session INTEGER DEFAULT 1
);

-- characters
CREATE TABLE IF NOT EXISTS characters (
  id            TEXT PRIMARY KEY,
  campaign_id   TEXT NOT NULL REFERENCES campaigns(id),
  name          TEXT NOT NULL,
  type          TEXT CHECK(type IN ('pc','npc')) DEFAULT 'pc',
  class         TEXT,
  race          TEXT,
  level         INTEGER DEFAULT 1,
  hp_current    INTEGER NOT NULL,
  hp_max        INTEGER NOT NULL,
  stat_str      INTEGER,
  stat_dex      INTEGER,
  stat_con      INTEGER,
  stat_int      INTEGER,
  stat_wis      INTEGER,
  stat_cha      INTEGER,
  gold          INTEGER DEFAULT 0,
  backstory     TEXT,
  inventory     TEXT,        -- JSON array
  conditions    TEXT,        -- JSON array
  relationships TEXT,        -- JSON object
  plot_flags    TEXT,        -- JSON array
  notes         TEXT,        -- DM-only private notes
  status        TEXT DEFAULT 'active',
  updated_at    TEXT DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(campaign_id, name)
);

-- plot_events
CREATE TABLE IF NOT EXISTS plot_events (
  id            TEXT PRIMARY KEY,
  campaign_id   TEXT NOT NULL REFERENCES campaigns(id),
  session       INTEGER NOT NULL,
  summary       TEXT NOT NULL,
  npcs_involved TEXT,        -- JSON array
  pcs_involved  TEXT,        -- JSON array
  consequences  TEXT,
  tags          TEXT,        -- JSON array
  created_at    TEXT DEFAULT CURRENT_TIMESTAMP
);

-- plot_hooks
CREATE TABLE IF NOT EXISTS plot_hooks (
  id             TEXT PRIMARY KEY,
  campaign_id    TEXT NOT NULL REFERENCES campaigns(id),
  hook           TEXT NOT NULL,
  session_opened INTEGER NOT NULL,
  event_id       TEXT REFERENCES plot_events(id),
  resolved       INTEGER DEFAULT 0,
  resolution     TEXT,
  resolved_at    TEXT
);

-- world_flags
CREATE TABLE IF NOT EXISTS world_flags (
  campaign_id TEXT NOT NULL REFERENCES campaigns(id),
  key         TEXT NOT NULL,
  value       TEXT,
  updated_at  TEXT DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (campaign_id, key)
);

-- roll_log
CREATE TABLE IF NOT EXISTS roll_log (
  id           TEXT PRIMARY KEY,
  campaign_id  TEXT NOT NULL REFERENCES campaigns(id),
  session      INTEGER,
  character    TEXT,
  notation     TEXT NOT NULL,
  total        INTEGER NOT NULL,
  rolls        TEXT NOT NULL,  -- JSON array of individual die results
  kept         TEXT,           -- JSON array after kh/kl
  modifier     INTEGER DEFAULT 0,
  reason       TEXT,
  advantage    INTEGER DEFAULT 0,
  disadvantage INTEGER DEFAULT 0,
  created_at   TEXT DEFAULT CURRENT_TIMESTAMP
);

-- sessions
CREATE TABLE IF NOT EXISTS sessions (
  campaign_id    TEXT NOT NULL REFERENCES campaigns(id),
  session        INTEGER NOT NULL,
  summary        TEXT,
  dm_notes       TEXT,
  hooks_opened   INTEGER DEFAULT 0,
  hooks_resolved INTEGER DEFAULT 0,
  created_at     TEXT DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (campaign_id, session)
);

-- checkpoints
CREATE TABLE IF NOT EXISTS checkpoints (
  id          TEXT PRIMARY KEY,
  campaign_id TEXT NOT NULL REFERENCES campaigns(id),
  session     INTEGER NOT NULL,
  note        TEXT NOT NULL,
  data        TEXT,        -- JSON turn data
  created_at  TEXT DEFAULT CURRENT_TIMESTAMP
);

-- indexes for common queries
CREATE INDEX IF NOT EXISTS idx_characters_campaign ON characters(campaign_id);
CREATE INDEX IF NOT EXISTS idx_plot_events_campaign_session ON plot_events(campaign_id, session);
CREATE INDEX IF NOT EXISTS idx_plot_hooks_campaign_resolved ON plot_hooks(campaign_id, resolved);
CREATE INDEX IF NOT EXISTS idx_roll_log_campaign_session ON roll_log(campaign_id, session);
CREATE INDEX IF NOT EXISTS idx_roll_log_character ON roll_log(character);
