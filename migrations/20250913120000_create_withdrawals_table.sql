-- +goose Up
-- Create ENUM types
CREATE TYPE payout_status AS ENUM ('new', 'sent', 'success', 'failed');
CREATE TYPE bank_type AS ENUM ('dummy', 'saman', 'mellat');

-- Create table
CREATE TABLE withdrawals (
    id UUID PRIMARY KEY,
    wallet_id UUID NOT NULL,
    status payout_status NOT NULL,
    bank bank_type NOT NULL,
    block_transaction_id BIGINT,
    withdrawal_transaction_id BIGINT,
    reverser_transaction_id BIGINT,
    amount BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_withdrawals_wallet_id ON withdrawals(wallet_id);
CREATE INDEX idx_withdrawals_status ON withdrawals(status);
CREATE INDEX idx_withdrawals_bank ON withdrawals(bank);
CREATE INDEX idx_withdrawals_block_tx_id ON withdrawals(block_transaction_id);
CREATE INDEX idx_withdrawals_withdrawal_tx_id ON withdrawals(withdrawal_transaction_id);
CREATE INDEX idx_withdrawals_reverser_tx_id ON withdrawals(reverser_transaction_id);

-- +goose Down
DROP TABLE withdrawals;

DROP TYPE payout_status;
DROP TYPE bank_type;
