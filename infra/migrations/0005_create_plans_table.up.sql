CREATE TABLE plans (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT UNIQUE NOT NULL,
  price_id TEXT UNIQUE NOT NULL,
  monthly_limit INT NOT NULL
);
