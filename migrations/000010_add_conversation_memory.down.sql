-- Remove conversation memory fields
DROP INDEX IF EXISTS idx_conversations_message_count;
ALTER TABLE conversations DROP COLUMN IF EXISTS message_count;
ALTER TABLE conversations DROP COLUMN IF EXISTS summary_up_to;
ALTER TABLE conversations DROP COLUMN IF EXISTS summary;
