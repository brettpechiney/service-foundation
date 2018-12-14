-- +migrate Up
CREATE TABLE IF NOT EXISTS workout (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	ishi_user_id UUID NOT NULL,
	note VARCHAR(50),
	created_by VARCHAR(20) NOT NULL,
	created_on TIMESTAMPTZ NOT NULL DEFAULT statement_timestamp(),
	modified_by VARCHAR(20) NOT NULL,
	modified_on TIMESTAMPTZ NOT NULL DEFAULT statement_timestamp(),
	INDEX (ishi_user_id)
);

-- +migrate Down
DROP TABLE workout;