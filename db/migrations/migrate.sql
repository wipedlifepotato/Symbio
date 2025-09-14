CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
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

CREATE TABLE IF NOT EXISTS transactions (
    id SERIAL PRIMARY KEY,
    from_wallet_id INT REFERENCES wallets(id) ON DELETE SET NULL,
    to_wallet_id INT REFERENCES wallets(id) ON DELETE SET NULL,
    to_address VARCHAR(255),
    amount NUMERIC(30,12) NOT NULL,
    currency VARCHAR(10) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
/*

CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    client_id INT REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(50),
    budget NUMERIC(20,8),
    currency VARCHAR(10) DEFAULT 'USD',
    status VARCHAR(20) DEFAULT 'open',
    created_at TIMESTAMP DEFAULT NOW(),
    deadline TIMESTAMP
);

CREATE TABLE IF NOT EXISTS task_offers (
    id SERIAL PRIMARY KEY,
    task_id INT REFERENCES tasks(id) ON DELETE CASCADE,
    freelancer_id INT REFERENCES users(id) ON DELETE CASCADE,
    price NUMERIC(20,8),
    message TEXT,
    status VARCHAR(20) DEFAULT 'pending', 
    created_at TIMESTAMP DEFAULT NOW()
);



CREATE TABLE IF NOT EXISTS messages (
    id SERIAL PRIMARY KEY,
    sender_id INT REFERENCES users(id) ON DELETE CASCADE,
    recipient_id INT REFERENCES users(id) ON DELETE CASCADE,
    task_id INT REFERENCES tasks(id) ON DELETE CASCADE,
    message TEXT NOT NULL,
    read BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);


*/