-- Add remaining_amount column to credits table
ALTER TABLE credits
ADD COLUMN remaining_amount DECIMAL(15,2) NOT NULL DEFAULT 0;

-- Update existing credits to set remaining_amount equal to amount
UPDATE credits SET remaining_amount = amount;

-- Add check constraint to ensure remaining_amount is not negative
ALTER TABLE credits
ADD CONSTRAINT credits_remaining_amount_check CHECK (remaining_amount >= 0);

-- +migrate Down
ALTER TABLE credits DROP COLUMN remaining_amount; 