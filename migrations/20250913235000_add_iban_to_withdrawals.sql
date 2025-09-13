-- +goose Up
ALTER TABLE withdrawals
    ADD COLUMN iban VARCHAR(24) NOT NULL DEFAULT '';

CREATE INDEX idx_withdrawals_iban ON withdrawals(iban);

-- +goose Down
DROP INDEX IF EXISTS idx_withdrawals_iban;

ALTER TABLE withdrawals
    DROP COLUMN iban;
