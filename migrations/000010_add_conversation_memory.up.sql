-- Add conversation memory fields for sliding window + summary approach
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS summary TEXT;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS summary_up_to UUID;
ALTER TABLE conversations ADD COLUMN IF NOT EXISTS message_count INT NOT NULL DEFAULT 0;

-- Index for finding conversations that need summary updates
CREATE INDEX idx_conversations_message_count ON conversations(message_count) WHERE deleted_at IS NULL AND summary IS NULL;

-- Add foreign key constraint for summary_up_to referencing messages
-- Note: We don't add a strict FK here because the referenced message might be deleted in future cleanup
COMMENT ON COLUMN conversations.summary IS 'Rolling summary of older conversation messages';
COMMENT ON COLUMN conversations.summary_up_to IS 'ID of the last message included in the summary';
COMMENT ON COLUMN conversations.message_count IS 'Total number of messages in the conversation';
