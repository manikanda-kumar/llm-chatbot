-- Initial schema for banking-core
-- This file is for reference; migrations are embedded in db.go

-- Customers
CREATE TABLE IF NOT EXISTS customers (
    id TEXT PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    phone TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Accounts
CREATE TABLE IF NOT EXISTS accounts (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL REFERENCES customers(id),
    account_type TEXT NOT NULL,
    balance REAL DEFAULT 0,
    currency TEXT DEFAULT 'CAD',
    status TEXT DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Transactions
CREATE TABLE IF NOT EXISTS transactions (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL REFERENCES accounts(id),
    type TEXT NOT NULL,
    amount REAL NOT NULL,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Transfers
CREATE TABLE IF NOT EXISTS transfers (
    id TEXT PRIMARY KEY,
    from_account_id TEXT NOT NULL REFERENCES accounts(id),
    to_account_id TEXT NOT NULL REFERENCES accounts(id),
    amount REAL NOT NULL,
    status TEXT DEFAULT 'pending',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Cards
CREATE TABLE IF NOT EXISTS cards (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL REFERENCES customers(id),
    account_id TEXT NOT NULL REFERENCES accounts(id),
    card_number TEXT NOT NULL,
    card_type TEXT NOT NULL,
    expiry_date TEXT NOT NULL,
    status TEXT DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Loans
CREATE TABLE IF NOT EXISTS loans (
    id TEXT PRIMARY KEY,
    customer_id TEXT NOT NULL REFERENCES customers(id),
    loan_type TEXT NOT NULL,
    principal REAL NOT NULL,
    interest_rate REAL NOT NULL,
    term_months INTEGER NOT NULL,
    monthly_payment REAL NOT NULL,
    balance REAL NOT NULL,
    status TEXT DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_accounts_customer ON accounts(customer_id);
CREATE INDEX IF NOT EXISTS idx_transactions_account ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_cards_customer ON cards(customer_id);
CREATE INDEX IF NOT EXISTS idx_loans_customer ON loans(customer_id);

-- Seed data
INSERT OR IGNORE INTO customers (id, first_name, last_name, email, phone) VALUES
    ('cust-001', 'John', 'Doe', 'john.doe@example.com', '+1-416-555-0101'),
    ('cust-002', 'Jane', 'Smith', 'jane.smith@example.com', '+1-416-555-0102');

INSERT OR IGNORE INTO accounts (id, customer_id, account_type, balance, currency, status) VALUES
    ('acc-001', 'cust-001', 'checking', 5000.00, 'CAD', 'active'),
    ('acc-002', 'cust-001', 'savings', 15000.00, 'CAD', 'active'),
    ('acc-003', 'cust-002', 'checking', 8500.00, 'CAD', 'active');

INSERT OR IGNORE INTO cards (id, customer_id, account_id, card_number, card_type, expiry_date, status) VALUES
    ('card-001', 'cust-001', 'acc-001', '****-****-****-1234', 'debit', '12/27', 'active'),
    ('card-002', 'cust-002', 'acc-003', '****-****-****-5678', 'credit', '06/26', 'active');

INSERT OR IGNORE INTO loans (id, customer_id, loan_type, principal, interest_rate, term_months, monthly_payment, balance, status) VALUES
    ('loan-001', 'cust-001', 'mortgage', 350000.00, 5.25, 300, 2100.50, 325000.00, 'active'),
    ('loan-002', 'cust-002', 'auto', 25000.00, 6.99, 60, 495.00, 18500.00, 'active');
