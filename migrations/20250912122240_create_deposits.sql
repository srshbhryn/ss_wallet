-- +goose Up
-- +goose StatementBegin

CREATE TABLE deposits (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    apply_at TIMESTAMPTZ NOT NULL,
    amount BIGINT NOT NULL,
    description VARCHAR(255),
    block_transaction_id BIGINT,
    apply_transaction_id BIGINT
);

CREATE INDEX idx_deposits_user_id ON deposits(user_id);
CREATE INDEX idx_deposits_created_at ON deposits(created_at);
CREATE INDEX idx_deposits_apply_at ON deposits(apply_at);
CREATE INDEX idx_deposits_block_transaction_id ON deposits(block_transaction_id);
CREATE INDEX idx_deposits_apply_transaction_id ON deposits(apply_transaction_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS deposits;

-- +goose StatementEnd
