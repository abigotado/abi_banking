-- Create cards table
CREATE TABLE IF NOT EXISTS cards (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
    card_number VARCHAR(16) NOT NULL UNIQUE,
    expiry_date VARCHAR(5) NOT NULL,
    cvv VARCHAR(3) NOT NULL,
    card_type VARCHAR(10) NOT NULL CHECK (card_type IN ('debit', 'credit')),
    status VARCHAR(10) NOT NULL CHECK (status IN ('active', 'blocked')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create index on card_number for faster lookups
CREATE INDEX IF NOT EXISTS idx_cards_card_number ON cards(card_number);

-- Create index on user_id for faster user card queries
CREATE INDEX IF NOT EXISTS idx_cards_user_id ON cards(user_id);

-- Create index on account_id for faster account card queries
CREATE INDEX IF NOT EXISTS idx_cards_account_id ON cards(account_id);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_cards_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_cards_updated_at
    BEFORE UPDATE ON cards
    FOR EACH ROW
    EXECUTE FUNCTION update_cards_updated_at(); 