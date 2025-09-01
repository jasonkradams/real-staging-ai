CREATE TYPE image_status AS ENUM ('queued','processing','ready','error');

CREATE TABLE images (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  original_url TEXT NOT NULL,
  staged_url TEXT,
  room_type TEXT,
  style TEXT,
  seed BIGINT,
  status image_status NOT NULL DEFAULT 'queued',
  error TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_images_project ON images(project_id);
