CREATE TABLE contact_submissions (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT        NOT NULL,
  company       TEXT,
  email         TEXT        NOT NULL,
  phone         TEXT,
  category      TEXT,
  message       TEXT        NOT NULL,
  status        TEXT        NOT NULL DEFAULT 'new'
                  CHECK (status IN ('new', 'read', 'replied', 'archived')),
  ip_address    INET,
  user_agent    TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_contact_submissions_status_created
  ON contact_submissions (status, created_at DESC);

CREATE INDEX idx_contact_submissions_email
  ON contact_submissions (email);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS TRIGGER LANGUAGE plpgsql AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_contact_submissions_updated_at ON contact_submissions;

CREATE TRIGGER trg_contact_submissions_updated_at
  BEFORE UPDATE ON contact_submissions
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
