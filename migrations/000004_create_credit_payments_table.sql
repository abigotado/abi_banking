-- Create credit_payments table
CREATE TABLE IF NOT EXISTS credit_payments (
    id SERIAL PRIMARY KEY,
    credit_id INTEGER NOT NULL REFERENCES credits(id) ON DELETE CASCADE,
    amount DECIMAL(15,2) NOT NULL,
    payment_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('scheduled', 'paid', 'missed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on credit_id for faster credit payment queries
CREATE INDEX IF NOT EXISTS idx_credit_payments_credit_id ON credit_payments(credit_id);

-- Create index on payment_date for faster payment scheduling queries
CREATE INDEX IF NOT EXISTS idx_credit_payments_payment_date ON credit_payments(payment_date);

-- Create index on status for faster status-based queries
CREATE INDEX IF NOT EXISTS idx_credit_payments_status ON credit_payments(status);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_credit_payments_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_credit_payments_updated_at
    BEFORE UPDATE ON credit_payments
    FOR EACH ROW
    EXECUTE FUNCTION update_credit_payments_updated_at(); 