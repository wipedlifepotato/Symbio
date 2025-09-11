CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    mnemonic TEXT NOT NULL UNIQUE, 
    password_hash TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT now()
);

CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(10) NOT NULL, 
    address TEXT NOT NULL UNIQUE,
    balance NUMERIC(20,8) DEFAULT 0
);
