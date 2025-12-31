package repository

import (
	"banking-core/internal/models"

	"github.com/google/uuid"
)

type AccountRepository struct {
	db *DB
}

func NewAccountRepository(db *DB) *AccountRepository {
	return &AccountRepository{db: db}
}

func (r *AccountRepository) GetByID(id string) (*models.Account, error) {
	var account models.Account
	err := r.db.Get(&account, "SELECT * FROM accounts WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *AccountRepository) GetByCustomerID(customerID string) ([]models.Account, error) {
	var accounts []models.Account
	err := r.db.Select(&accounts, "SELECT * FROM accounts WHERE customer_id = ? ORDER BY created_at DESC", customerID)
	return accounts, err
}

func (r *AccountRepository) Create(req *models.CreateAccountRequest) (*models.Account, error) {
	id := uuid.New().String()
	currency := req.Currency
	if currency == "" {
		currency = "CAD"
	}
	_, err := r.db.Exec(
		"INSERT INTO accounts (id, customer_id, account_type, balance, currency, status) VALUES (?, ?, ?, 0, ?, 'active')",
		id, req.CustomerID, req.AccountType, currency,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *AccountRepository) UpdateBalance(id string, newBalance float64) error {
	_, err := r.db.Exec("UPDATE accounts SET balance = ? WHERE id = ?", newBalance, id)
	return err
}

func (r *AccountRepository) GetBalance(id string) (*models.BalanceResponse, error) {
	account, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}
	return &models.BalanceResponse{
		AccountID: account.ID,
		Balance:   account.Balance,
		Currency:  account.Currency,
	}, nil
}
