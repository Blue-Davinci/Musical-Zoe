-- +goose Up

CREATE TABLE IF NOT EXISTS api_keys(
    api_key bytea PRIMARY KEY,
    user_id bigint NOT NULL REFERENCES users ON DELETE CASCADE,
    expiry timestamp(0) with time zone NOT NULL DEFAULT NOW() + INTERVAL '3 day',
    scope text NOT NULL
);

-- Create indexes for faster lookups
CREATE INDEX idx_api_keys_user_id ON api_keys (user_id);
-- scope
CREATE INDEX idx_api_keys_scope ON api_keys (scope);


-- +goose Down
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_scope;
DROP TABLE IF EXISTS api_keys;