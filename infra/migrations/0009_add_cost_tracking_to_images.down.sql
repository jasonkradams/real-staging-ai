-- Remove indexes
DROP INDEX IF EXISTS idx_images_model;
DROP INDEX IF EXISTS idx_images_project_cost;

-- Remove cost tracking columns
ALTER TABLE images 
  DROP COLUMN IF EXISTS replicate_prediction_id,
  DROP COLUMN IF EXISTS processing_time_ms,
  DROP COLUMN IF EXISTS model_used,
  DROP COLUMN IF EXISTS cost_usd;
