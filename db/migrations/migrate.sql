
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    mnemonic TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    blocked BOOLEAN NOT NULL DEFAULT FALSE,
    is_admin BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS profiles (
    user_id INT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    full_name VARCHAR(100),
    bio TEXT,
    skills JSONB,
    avatar TEXT,
    rating NUMERIC(3,2) DEFAULT 0,
    completed_tasks INT DEFAULT 0
);

CREATE TABLE IF NOT EXISTS wallets (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    currency VARCHAR(10) NOT NULL,
    address TEXT NOT NULL UNIQUE,
    balance NUMERIC(30,12) DEFAULT 0
);
CREATE TABLE IF NOT EXISTS tickets (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    admin_id INT REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'open',
    subject TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    additional_users_have_access INT[] DEFAULT '{}'
);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id SERIAL PRIMARY KEY,
    ticket_id INT REFERENCES tickets(id) ON DELETE CASCADE,
    sender_id INT REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS wallet_transactions (
    txid text NOT NULL,
    wallet_id integer,
    amount numeric(20,8),
    currency text,
    confirmed boolean,
    created_at timestamp without time zone DEFAULT now(),
    CONSTRAINT wallet_transactions_pkey PRIMARY KEY (txid),
    CONSTRAINT wallet_transactions_txid_unique UNIQUE (txid)
);
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    client_id INT REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50),
    budget NUMERIC(20,8),
    currency VARCHAR(10) DEFAULT 'BTC',
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT NOW(),
    deadline TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tasks_client_id ON tasks (client_id);

CREATE TABLE IF NOT EXISTS task_offers (
    id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(id) ON DELETE CASCADE,
    freelancer_id INT REFERENCES users(id) ON DELETE CASCADE,
    price NUMERIC(20,8),
    message TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    accepted BOOLEAN DEFAULT FALSE -- заказчик выбрал исполнителя
);


CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_wallet_id INT REFERENCES wallets(id) ON DELETE SET NULL,
    to_wallet_id INT REFERENCES wallets(id) ON DELETE SET NULL,
    to_address VARCHAR(255),
    task_id INT REFERENCES tasks(id) ON DELETE SET NULL,
    amount NUMERIC(30,12) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    confirmed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS escrow_balances (
    id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(id) ON DELETE CASCADE,
    client_id INT REFERENCES users(id),
    freelancer_id INT REFERENCES users(id),
    amount NUMERIC(30,12) DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'BTC',
    status VARCHAR(20) DEFAULT 'pending', -- pending, released, refunded
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS disputes (
    id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(id) ON DELETE CASCADE,
    opened_by INT REFERENCES users(id),
    assigned_admin INT REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(20) DEFAULT 'open', -- open, resolved
    resolution VARCHAR(20), -- client_won, freelancer_won
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS dispute_messages (
    id SERIAL PRIMARY KEY,
    dispute_id INT REFERENCES disputes(id) ON DELETE CASCADE,
    sender_id INT REFERENCES users(id),
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS reviews (
    id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(id) ON DELETE CASCADE,
    reviewer_id INT REFERENCES users(id),
    reviewed_id INT REFERENCES users(id), -- к кому отзыв
    rating INT CHECK (rating >= 1 AND rating <= 5),
    comment TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_task_offers_task_id ON task_offers (task_id);
CREATE INDEX IF NOT EXISTS idx_task_offers_freelancer_id ON task_offers (freelancer_id);

-- Chat rooms
CREATE TABLE IF NOT EXISTS chat_rooms (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Participants in chat rooms
CREATE TABLE IF NOT EXISTS chat_participants (
    id SERIAL PRIMARY KEY,
    chat_room_id INT REFERENCES chat_rooms(id) ON DELETE CASCADE,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    joined_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (chat_room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_participants_chat_room_id ON chat_participants (chat_room_id);
CREATE INDEX IF NOT EXISTS idx_chat_participants_user_id ON chat_participants (user_id);

-- Chat messages
CREATE TABLE IF NOT EXISTS chat_messages (
    id SERIAL PRIMARY KEY,
    chat_room_id INT REFERENCES chat_rooms(id) ON DELETE CASCADE,
    sender_id INT REFERENCES users(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_chat_room_id ON chat_messages (chat_room_id);
CREATE INDEX IF NOT EXISTS idx_chat_messages_sender_id ON chat_messages (sender_id);

-- Chat requests
CREATE TABLE IF NOT EXISTS chat_requests (
    id SERIAL PRIMARY KEY,
    requester_id INT REFERENCES users(id) ON DELETE CASCADE,
    requested_id INT REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'pending', -- pending, accepted, rejected
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE (requester_id, requested_id)
);

CREATE INDEX IF NOT EXISTS idx_chat_requests_requester_id ON chat_requests (requester_id);
CREATE INDEX IF NOT EXISTS idx_chat_requests_requested_id ON chat_requests (requested_id);

-- Add admin permissions and title
ALTER TABLE users ADD COLUMN IF NOT EXISTS admin_title VARCHAR(50) DEFAULT NULL;
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions INTEGER DEFAULT 0;

-- Add missing columns to transactions table
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS task_id INT REFERENCES tasks(id) ON DELETE SET NULL;
ALTER TABLE transactions ADD COLUMN IF NOT EXISTS confirmed BOOLEAN DEFAULT FALSE;
