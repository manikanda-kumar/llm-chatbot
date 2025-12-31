package repository

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sqlx.DB
}

func NewDB(dataSourceName string) (*DB, error) {
	db, err := sqlx.Connect("sqlite3", dataSourceName)
	if err != nil {
		return nil, err
	}

	// Enable foreign keys for SQLite
	_, err = db.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

func (db *DB) Migrate() error {
	schema := `
	-- Customers with extended fields for identity verification
	CREATE TABLE IF NOT EXISTS customers (
		id TEXT PRIMARY KEY,
		first_name TEXT NOT NULL,
		last_name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		phone TEXT,
		date_of_birth TEXT NOT NULL,
		gender TEXT NOT NULL,
		sin TEXT NOT NULL,
		address TEXT,
		city TEXT,
		province TEXT,
		postal_code TEXT,
		status TEXT DEFAULT 'active',
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

	-- Indexes for search optimization
	CREATE INDEX IF NOT EXISTS idx_accounts_customer ON accounts(customer_id);
	CREATE INDEX IF NOT EXISTS idx_transactions_account ON transactions(account_id);
	CREATE INDEX IF NOT EXISTS idx_cards_customer ON cards(customer_id);
	CREATE INDEX IF NOT EXISTS idx_loans_customer ON loans(customer_id);
	CREATE INDEX IF NOT EXISTS idx_customers_name ON customers(first_name, last_name);
	CREATE INDEX IF NOT EXISTS idx_customers_dob ON customers(date_of_birth);
	CREATE INDEX IF NOT EXISTS idx_customers_sin ON customers(sin);
	`

	_, err := db.Exec(schema)
	return err
}

func (db *DB) Seed() error {
	seed := `
	-- =====================================================
	-- 50 CUSTOMERS WITH REALISTIC OVERLAPS
	-- =====================================================
	-- Scenarios:
	-- 1. Same name, different DOB, different gender (John Smith x3)
	-- 2. Same name, same DOB, different gender (Michael Johnson x2)
	-- 3. Same DOB, different names
	-- 4. Inactive customers
	-- 5. Suspended customers
	-- =====================================================

	INSERT OR IGNORE INTO customers (id, first_name, last_name, email, phone, date_of_birth, gender, sin, address, city, province, postal_code, status) VALUES
		-- John Smith variations (same name, different people)
		('cust-001', 'John', 'Smith', 'john.smith1@email.com', '+1-416-555-0101', '1985-03-15', 'male', '***-***-001', '123 Main St', 'Toronto', 'ON', 'M5V 1A1', 'active'),
		('cust-002', 'John', 'Smith', 'john.smith2@email.com', '+1-416-555-0102', '1990-07-22', 'male', '***-***-002', '456 Oak Ave', 'Mississauga', 'ON', 'L5B 2C3', 'active'),
		('cust-003', 'John', 'Smith', 'johnsmith3@email.com', '+1-905-555-0103', '1978-11-08', 'male', '***-***-003', '789 Pine Rd', 'Brampton', 'ON', 'L6T 4K2', 'active'),

		-- Michael Johnson variations (same name, same DOB, different gender)
		('cust-004', 'Michael', 'Johnson', 'michael.j@email.com', '+1-416-555-0104', '1992-05-20', 'male', '***-***-004', '321 Elm St', 'Toronto', 'ON', 'M4W 1K5', 'active'),
		('cust-005', 'Michael', 'Johnson', 'michaelj@email.com', '+1-647-555-0105', '1992-05-20', 'female', '***-***-005', '654 Birch Lane', 'Scarborough', 'ON', 'M1B 3C4', 'active'),

		-- Same DOB cluster (1988-09-12)
		('cust-006', 'Sarah', 'Williams', 'sarah.w@email.com', '+1-416-555-0106', '1988-09-12', 'female', '***-***-006', '111 Queen St', 'Toronto', 'ON', 'M5H 2M9', 'active'),
		('cust-007', 'David', 'Brown', 'david.b@email.com', '+1-416-555-0107', '1988-09-12', 'male', '***-***-007', '222 King St', 'Toronto', 'ON', 'M5V 1J5', 'active'),
		('cust-008', 'Emma', 'Davis', 'emma.d@email.com', '+1-905-555-0108', '1988-09-12', 'female', '***-***-008', '333 Yonge St', 'North York', 'ON', 'M2N 5P9', 'active'),

		-- Inactive customers
		('cust-009', 'Robert', 'Wilson', 'robert.w@email.com', '+1-416-555-0109', '1975-02-28', 'male', '***-***-009', '444 Bay St', 'Toronto', 'ON', 'M5J 2T3', 'inactive'),
		('cust-010', 'Jennifer', 'Taylor', 'jennifer.t@email.com', '+1-416-555-0110', '1983-06-14', 'female', '***-***-010', '555 Bloor St', 'Toronto', 'ON', 'M4W 1A5', 'inactive'),

		-- Suspended customer
		('cust-011', 'William', 'Anderson', 'william.a@email.com', '+1-416-555-0111', '1995-12-01', 'male', '***-***-011', '666 Dundas St', 'Toronto', 'ON', 'M5T 1H4', 'suspended'),

		-- More active customers with diverse profiles
		('cust-012', 'Emily', 'Thomas', 'emily.t@email.com', '+1-416-555-0112', '1991-04-18', 'female', '***-***-012', '777 College St', 'Toronto', 'ON', 'M6G 1C2', 'active'),
		('cust-013', 'James', 'Jackson', 'james.j@email.com', '+1-647-555-0113', '1987-08-25', 'male', '***-***-013', '888 Spadina Ave', 'Toronto', 'ON', 'M5S 2H6', 'active'),
		('cust-014', 'Olivia', 'White', 'olivia.w@email.com', '+1-416-555-0114', '1993-01-30', 'female', '***-***-014', '999 Bathurst St', 'Toronto', 'ON', 'M5R 3G3', 'active'),
		('cust-015', 'Benjamin', 'Harris', 'benjamin.h@email.com', '+1-905-555-0115', '1980-10-05', 'male', '***-***-015', '100 Dufferin St', 'Toronto', 'ON', 'M6K 1Y9', 'active'),

		-- Another name cluster: Maria Garcia
		('cust-016', 'Maria', 'Garcia', 'maria.g1@email.com', '+1-416-555-0116', '1989-07-12', 'female', '***-***-016', '200 Ossington Ave', 'Toronto', 'ON', 'M6J 2Z8', 'active'),
		('cust-017', 'Maria', 'Garcia', 'maria.garcia@email.com', '+1-647-555-0117', '1994-03-28', 'female', '***-***-017', '300 Lansdowne Ave', 'Toronto', 'ON', 'M6K 2W4', 'active'),

		-- Same DOB cluster (1995-11-15)
		('cust-018', 'Alexander', 'Martinez', 'alex.m@email.com', '+1-416-555-0118', '1995-11-15', 'male', '***-***-018', '400 Roncesvalles Ave', 'Toronto', 'ON', 'M6R 2M8', 'active'),
		('cust-019', 'Sophia', 'Robinson', 'sophia.r@email.com', '+1-905-555-0119', '1995-11-15', 'female', '***-***-019', '500 Queen St W', 'Toronto', 'ON', 'M5V 2B4', 'active'),
		('cust-020', 'Daniel', 'Clark', 'daniel.c@email.com', '+1-416-555-0120', '1995-11-15', 'male', '***-***-020', '600 King St W', 'Toronto', 'ON', 'M5V 1M3', 'active'),

		-- More customers
		('cust-021', 'Isabella', 'Rodriguez', 'isabella.r@email.com', '+1-416-555-0121', '1986-04-22', 'female', '***-***-021', '700 Adelaide St', 'Toronto', 'ON', 'M5V 1R9', 'active'),
		('cust-022', 'Matthew', 'Lewis', 'matthew.l@email.com', '+1-647-555-0122', '1982-09-08', 'male', '***-***-022', '800 Richmond St', 'Toronto', 'ON', 'M5V 1Y3', 'active'),
		('cust-023', 'Ava', 'Lee', 'ava.lee@email.com', '+1-416-555-0123', '1997-02-14', 'female', '***-***-023', '900 Wellington St', 'Toronto', 'ON', 'M5V 1E6', 'active'),
		('cust-024', 'Christopher', 'Walker', 'chris.w@email.com', '+1-905-555-0124', '1979-06-30', 'male', '***-***-024', '1000 Front St', 'Toronto', 'ON', 'M5V 3B5', 'active'),
		('cust-025', 'Mia', 'Hall', 'mia.hall@email.com', '+1-416-555-0125', '1991-12-18', 'female', '***-***-025', '1100 Lakeshore Blvd', 'Toronto', 'ON', 'M8V 1A1', 'active'),

		-- Another inactive cluster
		('cust-026', 'Ethan', 'Allen', 'ethan.a@email.com', '+1-416-555-0126', '1988-03-05', 'male', '***-***-026', '1200 Marine Parade', 'Toronto', 'ON', 'M8V 1B2', 'inactive'),
		('cust-027', 'Charlotte', 'Young', 'charlotte.y@email.com', '+1-647-555-0127', '1993-08-21', 'female', '***-***-027', '1300 Park Lawn Rd', 'Toronto', 'ON', 'M8V 3C4', 'inactive'),

		-- David Lee variations
		('cust-028', 'David', 'Lee', 'david.lee1@email.com', '+1-416-555-0128', '1984-05-17', 'male', '***-***-028', '1400 The Queensway', 'Toronto', 'ON', 'M8Z 1S7', 'active'),
		('cust-029', 'David', 'Lee', 'david.lee2@email.com', '+1-905-555-0129', '1990-10-03', 'male', '***-***-029', '1500 Islington Ave', 'Toronto', 'ON', 'M9A 3M9', 'active'),

		-- More diverse customers
		('cust-030', 'Amelia', 'King', 'amelia.k@email.com', '+1-416-555-0130', '1996-01-25', 'female', '***-***-030', '1600 Kipling Ave', 'Toronto', 'ON', 'M9V 3T1', 'active'),
		('cust-031', 'Andrew', 'Wright', 'andrew.w@email.com', '+1-647-555-0131', '1981-07-11', 'male', '***-***-031', '1700 Martin Grove Rd', 'Toronto', 'ON', 'M9V 4W3', 'active'),
		('cust-032', 'Harper', 'Lopez', 'harper.l@email.com', '+1-416-555-0132', '1994-11-28', 'female', '***-***-032', '1800 Albion Rd', 'Toronto', 'ON', 'M9V 1B6', 'active'),
		('cust-033', 'Joshua', 'Hill', 'joshua.h@email.com', '+1-905-555-0133', '1987-04-09', 'male', '***-***-033', '1900 Finch Ave W', 'Toronto', 'ON', 'M9M 2K1', 'active'),
		('cust-034', 'Evelyn', 'Scott', 'evelyn.s@email.com', '+1-416-555-0134', '1992-06-16', 'female', '***-***-034', '2000 Jane St', 'Toronto', 'ON', 'M9N 2C8', 'active'),
		('cust-035', 'Logan', 'Green', 'logan.g@email.com', '+1-647-555-0135', '1985-09-23', 'male', '***-***-035', '2100 Weston Rd', 'Toronto', 'ON', 'M9N 1Z4', 'active'),

		-- Same name, same DOB: Sarah Williams (duplicate risk scenario)
		('cust-036', 'Sarah', 'Williams', 'sarahw@email.com', '+1-905-555-0136', '1988-09-12', 'female', '***-***-036', '2200 Lawrence Ave W', 'Toronto', 'ON', 'M9N 1J6', 'active'),

		-- More customers
		('cust-037', 'Ryan', 'Adams', 'ryan.a@email.com', '+1-416-555-0137', '1983-02-07', 'male', '***-***-037', '2300 Keele St', 'Toronto', 'ON', 'M3M 2H2', 'active'),
		('cust-038', 'Lily', 'Nelson', 'lily.n@email.com', '+1-416-555-0138', '1998-05-31', 'female', '***-***-038', '2400 Dufferin St N', 'Toronto', 'ON', 'M3H 5R8', 'active'),
		('cust-039', 'Nathan', 'Carter', 'nathan.c@email.com', '+1-647-555-0139', '1986-08-14', 'male', '***-***-039', '2500 Bathurst St N', 'Toronto', 'ON', 'M3H 3P7', 'active'),
		('cust-040', 'Grace', 'Mitchell', 'grace.m@email.com', '+1-905-555-0140', '1990-12-02', 'female', '***-***-040', '2600 Yonge St N', 'Toronto', 'ON', 'M4P 2J4', 'active'),

		-- Customers with only loans (no accounts)
		('cust-041', 'Lucas', 'Perez', 'lucas.p@email.com', '+1-416-555-0141', '1989-03-19', 'male', '***-***-041', '2700 Bayview Ave', 'Toronto', 'ON', 'M4G 3C1', 'active'),
		('cust-042', 'Chloe', 'Roberts', 'chloe.r@email.com', '+1-416-555-0142', '1995-07-26', 'female', '***-***-042', '2800 Leslie St', 'Toronto', 'ON', 'M4M 3E2', 'active'),

		-- Senior customers
		('cust-043', 'George', 'Turner', 'george.t@email.com', '+1-416-555-0143', '1955-01-10', 'male', '***-***-043', '2900 Don Mills Rd', 'Toronto', 'ON', 'M3B 2W6', 'active'),
		('cust-044', 'Margaret', 'Phillips', 'margaret.p@email.com', '+1-905-555-0144', '1960-04-25', 'female', '***-***-044', '3000 Victoria Park Ave', 'Toronto', 'ON', 'M4A 1A1', 'active'),

		-- Young customers
		('cust-045', 'Jack', 'Campbell', 'jack.c@email.com', '+1-647-555-0145', '2000-06-15', 'male', '***-***-045', '3100 Warden Ave', 'Toronto', 'ON', 'M1T 2T6', 'active'),
		('cust-046', 'Zoe', 'Parker', 'zoe.p@email.com', '+1-416-555-0146', '2001-09-08', 'female', '***-***-046', '3200 Birchmount Rd', 'Toronto', 'ON', 'M1P 2E3', 'active'),

		-- More name variations
		('cust-047', 'Henry', 'Evans', 'henry.e@email.com', '+1-416-555-0147', '1977-11-22', 'male', '***-***-047', '3300 Kennedy Rd', 'Toronto', 'ON', 'M1V 1E7', 'active'),
		('cust-048', 'Victoria', 'Edwards', 'victoria.e@email.com', '+1-905-555-0148', '1988-02-16', 'female', '***-***-048', '3400 McCowan Rd', 'Toronto', 'ON', 'M1S 2Y3', 'active'),
		('cust-049', 'Sebastian', 'Collins', 'sebastian.c@email.com', '+1-647-555-0149', '1993-10-30', 'male', '***-***-049', '3500 Markham Rd', 'Toronto', 'ON', 'M1B 5M7', 'active'),
		('cust-050', 'Penelope', 'Stewart', 'penelope.s@email.com', '+1-416-555-0150', '1991-05-12', 'female', '***-***-050', '3600 Ellesmere Rd', 'Toronto', 'ON', 'M1G 3T8', 'active');

	-- =====================================================
	-- ACCOUNTS - Various combinations
	-- =====================================================
	INSERT OR IGNORE INTO accounts (id, customer_id, account_type, balance, currency, status) VALUES
		-- Customer 001: All account types
		('acc-001', 'cust-001', 'checking', 5250.75, 'CAD', 'active'),
		('acc-002', 'cust-001', 'savings', 25000.00, 'CAD', 'active'),
		('acc-003', 'cust-001', 'credit', -1500.00, 'CAD', 'active'),

		-- Customer 002: Checking and savings only
		('acc-004', 'cust-002', 'checking', 3200.50, 'CAD', 'active'),
		('acc-005', 'cust-002', 'savings', 12500.00, 'CAD', 'active'),

		-- Customer 003: Only checking, one inactive
		('acc-006', 'cust-003', 'checking', 8750.25, 'CAD', 'active'),
		('acc-007', 'cust-003', 'checking', 0.00, 'CAD', 'inactive'),

		-- Customer 004: Multiple savings
		('acc-008', 'cust-004', 'savings', 50000.00, 'CAD', 'active'),
		('acc-009', 'cust-004', 'savings', 15000.00, 'CAD', 'active'),
		('acc-010', 'cust-004', 'checking', 2100.00, 'CAD', 'active'),

		-- Customer 005: Credit only
		('acc-011', 'cust-005', 'credit', -3500.00, 'CAD', 'active'),

		-- Customer 006-008: Standard accounts
		('acc-012', 'cust-006', 'checking', 4500.00, 'CAD', 'active'),
		('acc-013', 'cust-006', 'savings', 18000.00, 'CAD', 'active'),
		('acc-014', 'cust-007', 'checking', 6200.00, 'CAD', 'active'),
		('acc-015', 'cust-008', 'checking', 3800.00, 'CAD', 'active'),
		('acc-016', 'cust-008', 'savings', 22000.00, 'CAD', 'active'),

		-- Inactive customers (cust-009, 010): All accounts frozen
		('acc-017', 'cust-009', 'checking', 1500.00, 'CAD', 'frozen'),
		('acc-018', 'cust-009', 'savings', 5000.00, 'CAD', 'frozen'),
		('acc-019', 'cust-010', 'checking', 2200.00, 'CAD', 'frozen'),

		-- Suspended customer (cust-011): Mixed status
		('acc-020', 'cust-011', 'checking', 500.00, 'CAD', 'frozen'),
		('acc-021', 'cust-011', 'savings', 8000.00, 'CAD', 'frozen'),

		-- More customers with varied accounts
		('acc-022', 'cust-012', 'checking', 7500.00, 'CAD', 'active'),
		('acc-023', 'cust-012', 'savings', 35000.00, 'CAD', 'active'),
		('acc-024', 'cust-012', 'credit', -2000.00, 'CAD', 'active'),

		('acc-025', 'cust-013', 'checking', 4100.00, 'CAD', 'active'),
		('acc-026', 'cust-014', 'checking', 5600.00, 'CAD', 'active'),
		('acc-027', 'cust-014', 'savings', 15500.00, 'CAD', 'inactive'),

		('acc-028', 'cust-015', 'checking', 9800.00, 'CAD', 'active'),
		('acc-029', 'cust-015', 'savings', 45000.00, 'CAD', 'active'),

		('acc-030', 'cust-016', 'checking', 3300.00, 'CAD', 'active'),
		('acc-031', 'cust-017', 'checking', 2800.00, 'CAD', 'active'),
		('acc-032', 'cust-017', 'savings', 11000.00, 'CAD', 'active'),

		('acc-033', 'cust-018', 'checking', 6700.00, 'CAD', 'active'),
		('acc-034', 'cust-019', 'checking', 4900.00, 'CAD', 'active'),
		('acc-035', 'cust-020', 'checking', 5100.00, 'CAD', 'active'),

		('acc-036', 'cust-021', 'checking', 8200.00, 'CAD', 'active'),
		('acc-037', 'cust-021', 'savings', 28000.00, 'CAD', 'active'),

		('acc-038', 'cust-022', 'checking', 12500.00, 'CAD', 'active'),
		('acc-039', 'cust-023', 'checking', 1800.00, 'CAD', 'active'),
		('acc-040', 'cust-024', 'checking', 15000.00, 'CAD', 'active'),
		('acc-041', 'cust-024', 'savings', 75000.00, 'CAD', 'active'),

		('acc-042', 'cust-025', 'checking', 3600.00, 'CAD', 'active'),

		-- Inactive customers accounts
		('acc-043', 'cust-026', 'checking', 900.00, 'CAD', 'inactive'),
		('acc-044', 'cust-027', 'checking', 1200.00, 'CAD', 'inactive'),

		('acc-045', 'cust-028', 'checking', 7100.00, 'CAD', 'active'),
		('acc-046', 'cust-029', 'checking', 4400.00, 'CAD', 'active'),
		('acc-047', 'cust-029', 'savings', 19000.00, 'CAD', 'active'),

		('acc-048', 'cust-030', 'checking', 2600.00, 'CAD', 'active'),
		('acc-049', 'cust-031', 'checking', 11200.00, 'CAD', 'active'),
		('acc-050', 'cust-032', 'checking', 5800.00, 'CAD', 'active'),

		('acc-051', 'cust-033', 'checking', 6400.00, 'CAD', 'active'),
		('acc-052', 'cust-034', 'checking', 3900.00, 'CAD', 'active'),
		('acc-053', 'cust-034', 'savings', 16500.00, 'CAD', 'active'),

		('acc-054', 'cust-035', 'checking', 8600.00, 'CAD', 'active'),
		('acc-055', 'cust-036', 'checking', 4700.00, 'CAD', 'active'),
		('acc-056', 'cust-037', 'checking', 5300.00, 'CAD', 'active'),
		('acc-057', 'cust-038', 'checking', 1900.00, 'CAD', 'active'),
		('acc-058', 'cust-039', 'checking', 7800.00, 'CAD', 'active'),
		('acc-059', 'cust-040', 'checking', 4200.00, 'CAD', 'active'),
		('acc-060', 'cust-040', 'savings', 21000.00, 'CAD', 'active'),

		-- Customers 041-042: LOAN ONLY (no regular accounts, just loans below)

		-- Senior customers with substantial savings
		('acc-061', 'cust-043', 'checking', 8500.00, 'CAD', 'active'),
		('acc-062', 'cust-043', 'savings', 150000.00, 'CAD', 'active'),
		('acc-063', 'cust-044', 'checking', 6200.00, 'CAD', 'active'),
		('acc-064', 'cust-044', 'savings', 95000.00, 'CAD', 'active'),

		-- Young customers with minimal accounts
		('acc-065', 'cust-045', 'checking', 850.00, 'CAD', 'active'),
		('acc-066', 'cust-046', 'checking', 1200.00, 'CAD', 'active'),

		('acc-067', 'cust-047', 'checking', 9100.00, 'CAD', 'active'),
		('acc-068', 'cust-047', 'savings', 32000.00, 'CAD', 'active'),

		('acc-069', 'cust-048', 'checking', 5500.00, 'CAD', 'active'),
		('acc-070', 'cust-049', 'checking', 4000.00, 'CAD', 'active'),
		('acc-071', 'cust-050', 'checking', 6800.00, 'CAD', 'active'),
		('acc-072', 'cust-050', 'savings', 24000.00, 'CAD', 'active');

	-- =====================================================
	-- CARDS
	-- =====================================================
	INSERT OR IGNORE INTO cards (id, customer_id, account_id, card_number, card_type, expiry_date, status) VALUES
		-- Customer 001: Multiple cards
		('card-001', 'cust-001', 'acc-001', '****-****-****-1001', 'debit', '12/27', 'active'),
		('card-002', 'cust-001', 'acc-003', '****-****-****-1002', 'credit', '06/28', 'active'),
		('card-003', 'cust-001', 'acc-001', '****-****-****-1003', 'debit', '03/25', 'expired'),

		-- Customer 002
		('card-004', 'cust-002', 'acc-004', '****-****-****-2001', 'debit', '09/27', 'active'),

		-- Customer 003: One blocked card
		('card-005', 'cust-003', 'acc-006', '****-****-****-3001', 'debit', '11/26', 'active'),
		('card-006', 'cust-003', 'acc-006', '****-****-****-3002', 'debit', '05/27', 'blocked'),

		-- Customer 004-008
		('card-007', 'cust-004', 'acc-010', '****-****-****-4001', 'debit', '08/28', 'active'),
		('card-008', 'cust-005', 'acc-011', '****-****-****-5001', 'credit', '02/27', 'active'),
		('card-009', 'cust-006', 'acc-012', '****-****-****-6001', 'debit', '07/27', 'active'),
		('card-010', 'cust-007', 'acc-014', '****-****-****-7001', 'debit', '10/28', 'active'),
		('card-011', 'cust-008', 'acc-015', '****-****-****-8001', 'debit', '04/27', 'active'),

		-- Frozen cards for inactive/suspended customers
		('card-012', 'cust-009', 'acc-017', '****-****-****-9001', 'debit', '01/26', 'blocked'),
		('card-013', 'cust-010', 'acc-019', '****-****-****-1010', 'debit', '06/26', 'blocked'),
		('card-014', 'cust-011', 'acc-020', '****-****-****-1011', 'debit', '09/27', 'blocked'),

		-- More active cards
		('card-015', 'cust-012', 'acc-022', '****-****-****-1201', 'debit', '12/28', 'active'),
		('card-016', 'cust-012', 'acc-024', '****-****-****-1202', 'credit', '03/27', 'active'),
		('card-017', 'cust-013', 'acc-025', '****-****-****-1301', 'debit', '05/28', 'active'),
		('card-018', 'cust-014', 'acc-026', '****-****-****-1401', 'debit', '08/27', 'active'),
		('card-019', 'cust-015', 'acc-028', '****-****-****-1501', 'debit', '11/28', 'active'),
		('card-020', 'cust-016', 'acc-030', '****-****-****-1601', 'debit', '02/28', 'active'),
		('card-021', 'cust-017', 'acc-031', '****-****-****-1701', 'debit', '06/27', 'active'),
		('card-022', 'cust-018', 'acc-033', '****-****-****-1801', 'debit', '09/28', 'active'),
		('card-023', 'cust-019', 'acc-034', '****-****-****-1901', 'debit', '12/27', 'active'),
		('card-024', 'cust-020', 'acc-035', '****-****-****-2001', 'debit', '04/28', 'active'),

		-- More cards
		('card-025', 'cust-021', 'acc-036', '****-****-****-2101', 'debit', '07/27', 'active'),
		('card-026', 'cust-022', 'acc-038', '****-****-****-2201', 'debit', '10/28', 'active'),
		('card-027', 'cust-024', 'acc-040', '****-****-****-2401', 'debit', '01/28', 'active'),
		('card-028', 'cust-028', 'acc-045', '****-****-****-2801', 'debit', '05/27', 'active'),
		('card-029', 'cust-029', 'acc-046', '****-****-****-2901', 'debit', '08/28', 'active'),
		('card-030', 'cust-043', 'acc-061', '****-****-****-4301', 'debit', '11/27', 'active'),
		('card-031', 'cust-044', 'acc-063', '****-****-****-4401', 'debit', '03/28', 'active'),
		('card-032', 'cust-045', 'acc-065', '****-****-****-4501', 'debit', '06/29', 'active'),
		('card-033', 'cust-046', 'acc-066', '****-****-****-4601', 'debit', '09/29', 'active'),
		('card-034', 'cust-047', 'acc-067', '****-****-****-4701', 'debit', '12/28', 'active'),
		('card-035', 'cust-050', 'acc-071', '****-****-****-5001', 'debit', '04/28', 'active');

	-- =====================================================
	-- LOANS - Various types and statuses
	-- =====================================================
	INSERT OR IGNORE INTO loans (id, customer_id, loan_type, principal, interest_rate, term_months, monthly_payment, balance, status) VALUES
		-- Mortgages
		('loan-001', 'cust-001', 'mortgage', 450000.00, 5.25, 300, 2700.50, 425000.00, 'active'),
		('loan-002', 'cust-015', 'mortgage', 650000.00, 4.99, 300, 3800.25, 580000.00, 'active'),
		('loan-003', 'cust-024', 'mortgage', 380000.00, 5.49, 240, 2650.00, 320000.00, 'active'),
		('loan-004', 'cust-043', 'mortgage', 250000.00, 4.75, 180, 2100.00, 45000.00, 'active'),
		('loan-005', 'cust-047', 'mortgage', 520000.00, 5.15, 300, 3100.00, 495000.00, 'active'),

		-- Auto loans
		('loan-006', 'cust-002', 'auto', 35000.00, 6.49, 60, 685.00, 28000.00, 'active'),
		('loan-007', 'cust-004', 'auto', 42000.00, 5.99, 72, 695.00, 36500.00, 'active'),
		('loan-008', 'cust-012', 'auto', 28000.00, 6.99, 48, 670.00, 18500.00, 'active'),
		('loan-009', 'cust-022', 'auto', 55000.00, 5.49, 60, 1050.00, 42000.00, 'active'),
		('loan-010', 'cust-035', 'auto', 32000.00, 6.25, 60, 625.00, 24000.00, 'active'),

		-- Personal loans
		('loan-011', 'cust-006', 'personal', 15000.00, 8.99, 36, 475.00, 11000.00, 'active'),
		('loan-012', 'cust-013', 'personal', 25000.00, 9.49, 48, 625.00, 20000.00, 'active'),
		('loan-013', 'cust-019', 'personal', 10000.00, 7.99, 24, 455.00, 6500.00, 'active'),
		('loan-014', 'cust-025', 'personal', 18000.00, 8.49, 36, 565.00, 14000.00, 'active'),
		('loan-015', 'cust-033', 'personal', 8000.00, 9.99, 24, 365.00, 5200.00, 'active'),

		-- Loan-only customers (no regular accounts)
		('loan-016', 'cust-041', 'mortgage', 320000.00, 5.35, 300, 1950.00, 305000.00, 'active'),
		('loan-017', 'cust-041', 'auto', 28000.00, 6.75, 60, 555.00, 22000.00, 'active'),
		('loan-018', 'cust-042', 'personal', 20000.00, 8.25, 48, 490.00, 16500.00, 'active'),

		-- Paid off loans
		('loan-019', 'cust-044', 'auto', 30000.00, 5.99, 60, 580.00, 0.00, 'paid_off'),
		('loan-020', 'cust-031', 'personal', 12000.00, 8.99, 36, 380.00, 0.00, 'paid_off'),

		-- Defaulted loan
		('loan-021', 'cust-009', 'personal', 15000.00, 9.99, 36, 485.00, 12500.00, 'defaulted'),

		-- More active loans
		('loan-022', 'cust-007', 'auto', 38000.00, 6.49, 60, 745.00, 31000.00, 'active'),
		('loan-023', 'cust-021', 'mortgage', 410000.00, 5.09, 300, 2450.00, 385000.00, 'active'),
		('loan-024', 'cust-028', 'personal', 22000.00, 8.75, 48, 545.00, 17500.00, 'active'),
		('loan-025', 'cust-034', 'auto', 45000.00, 5.75, 72, 740.00, 39000.00, 'active'),

		-- Student-age customers with small loans
		('loan-026', 'cust-045', 'personal', 5000.00, 7.49, 24, 225.00, 3800.00, 'active'),
		('loan-027', 'cust-046', 'personal', 3500.00, 7.99, 18, 210.00, 2400.00, 'active'),

		-- Senior with paid off mortgage
		('loan-028', 'cust-043', 'personal', 8000.00, 6.99, 24, 355.00, 4500.00, 'active'),

		-- More variety
		('loan-029', 'cust-050', 'auto', 33000.00, 6.25, 60, 645.00, 27500.00, 'active'),
		('loan-030', 'cust-037', 'mortgage', 285000.00, 5.45, 240, 2050.00, 265000.00, 'active');
	`
	_, err := db.Exec(seed)
	return err
}
