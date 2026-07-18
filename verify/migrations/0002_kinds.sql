-- DT-22: verification kinds. 'atom' = a reflow observation hash (v0 behavior);
-- 'c2pa_url' = fetch the URL, validate C2PA Content Credentials, and check the
-- asset bytes against the observation log in the same verdict.
ALTER TABLE verdicts ADD COLUMN kind TEXT NOT NULL DEFAULT 'atom';
ALTER TABLE verdicts ADD COLUMN url TEXT;
