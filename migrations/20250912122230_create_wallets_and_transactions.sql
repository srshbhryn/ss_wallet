-- +goose Up
-- +goose StatementBegin

CREATE TABLE wallets (
    user_id UUID PRIMARY KEY,
    available_balance BIGINT NOT NULL DEFAULT 0,
    blocked_balance BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    id BIGSERIAL PRIMARY KEY,
    wallet_id UUID NOT NULL REFERENCES wallets(user_id) ON DELETE CASCADE,
    amount BIGINT NOT NULL DEFAULT 0,
    blocked_amount BIGINT NOT NULL DEFAULT 0,
    reference UUID NOT NULL,
    description VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_transactions_wallet_id ON transactions(wallet_id);
CREATE INDEX idx_transactions_reference ON transactions(reference);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS wallets;

-- +goose StatementEnd
