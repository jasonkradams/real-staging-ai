-- Add cost tracking columns to images table
ALTER TABLE images 
  ADD COLUMN cost_usd DECIMAL(10, 4) DEFAULT 0.00,
  ADD COLUMN model_used VARCHAR(255) CHECK (model_used ~ '^[a-zA-Z0-9_-]*$'),
  ADD COLUMN processing_time_ms INTEGER,
  ADD COLUMN replicate_prediction_id VARCHAR(255) CHECK (replicate_prediction_id ~ '^[a-zA-Z0-9_-]*$');

-- Add index for cost aggregation queries
CREATE INDEX idx_images_project_cost ON images(project_id, cost_usd);
CREATE INDEX idx_images_model ON images(model_used);

-- Add comments
COMMENT ON COLUMN images.cost_usd IS 'Cost in USD for processing this image';
COMMENT ON COLUMN images.model_used IS 'The AI model ID used to process this image';
COMMENT ON COLUMN images.processing_time_ms IS 'Processing time in milliseconds';
COMMENT ON COLUMN images.replicate_prediction_id IS 'Replicate prediction ID for tracking and billing';

