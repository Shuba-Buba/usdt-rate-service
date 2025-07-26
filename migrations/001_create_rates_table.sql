
CREATE TABLE IF NOT EXISTS rates (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    data JSONB NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_rates_timestamp ON rates(timestamp);
