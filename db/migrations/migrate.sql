CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    mnemonic TEXT NOT NULL UNIQUE, 
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    blocked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(10) NOT NULL, 
    address TEXT NOT NULL UNIQUE,
    balance NUMERIC(30,12) DEFAULT 0
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    txid TEXT PRIMARY KEY,
    wallet_id INT,
    amount NUMERIC(20,8),
    currency TEXT,
    confirmed BOOL,
    created_at TIMESTAMP DEFAULT NOW()
);


