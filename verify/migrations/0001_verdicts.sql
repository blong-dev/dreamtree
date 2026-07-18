-- The lazy index: only hashes someone actually asked about.
CREATE TABLE IF NOT EXISTS verdicts (
  hash TEXT PRIMARY KEY,
  status TEXT NOT NULL,             -- pending | observed | converged | not_found
  verdict TEXT,                     -- full JSON verdict (NULL while pending)
  requested_at TEXT NOT NULL DEFAULT (datetime('now')),
  resolved_at TEXT
);
CREATE INDEX IF NOT EXISTS idx_verdicts_status ON verdicts(status);
