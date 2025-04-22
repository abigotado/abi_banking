-- Create credits table
CREATE TABLE IF NOT EXISTS credits (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL,
    interest_rate DECIMAL(5,2) NOT NULL,
    term_months INTEGER NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'paid', 'defaulted')),
    remaining_amount DECIMAL(15,2) NOT NULL,
    next_payment_date TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on user_id for faster user credit queries
CREATE INDEX IF NOT EXISTS idx_credits_user_id ON credits(user_id);

-- Create index on account_id for faster account credit queries
CREATE INDEX IF NOT EXISTS idx_credits_account_id ON credits(account_id);

-- Create index on status for faster status-based queries
CREATE INDEX IF NOT EXISTS idx_credits_status ON credits(status);

-- Create index on next_payment_date for faster payment scheduling queries
CREATE INDEX IF NOT EXISTS idx_credits_next_payment_date ON credits(next_payment_date);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_credits_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_credits_updated_at
    BEFORE UPDATE ON credits
    FOR EACH ROW
    EXECUTE FUNCTION update_credits_updated_at(); 