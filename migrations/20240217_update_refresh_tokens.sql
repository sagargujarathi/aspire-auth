-- Make service_id nullable
ALTER TABLE refresh_tokens ALTER COLUMN service_id DROP NOT NULL;
