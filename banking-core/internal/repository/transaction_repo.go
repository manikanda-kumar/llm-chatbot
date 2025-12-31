package repository

import (
	"errors"

	"banking-core/internal/models"

	"github.com/google/uuid"
)

type TransactionRepository struct {
	db         *DB
	accountRepo *AccountRepository
}

func NewTransactionRepository(db *DB, accountRepo *AccountRepository) *TransactionRepository {
	return &TransactionRepository{db: db, accountRepo: accountRepo}
}

func (r *TransactionRepository) GetByID(id string) (*models.Transaction, error) {
	var txn models.Transaction
	err := r.db.Get(&txn, "SELECT * FROM transactions WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &txn, nil
}

func (r *TransactionRepository) GetByAccountID(accountID string, limit, offset int) ([]models.Transaction, error) {
	var transactions []models.Transaction
	err := r.db.Select(&transactions,
		"SELECT * FROM transactions WHERE account_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
		accountID, limit, offset,
	)
	return transactions, err
}

func (r *TransactionRepository) Create(req *models.CreateTransactionRequest) (*models.Transaction, error) {
	// Get current account balance
	account, err := r.accountRepo.GetByID(req.AccountID)
	if err != nil {
		return nil, err
	}

	var newBalance float64
	switch req.Type {
	case models.TransactionTypeDeposit, models.TransactionTypeTransferIn:
		newBalance = account.Balance + req.Amount
	case models.TransactionTypeWithdrawal, models.TransactionTypeTransferOut:
		if account.Balance < req.Amount {
			return nil, errors.New("insufficient funds")
		}
		newBalance = account.Balance - req.Amount
	default:
		return nil, errors.New("invalid transaction type")
	}

	// Start transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create transaction record
	id := uuid.New().String()
	_, err = tx.Exec(
		"INSERT INTO transactions (id, account_id, type, amount, description) VALUES (?, ?, ?, ?, ?)",
		id, req.AccountID, req.Type, req.Amount, req.Description,
	)
	if err != nil {
		return nil, err
	}

	// Update account balance
	_, err = tx.Exec("UPDATE accounts SET balance = ? WHERE id = ?", newBalance, req.AccountID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(id)
}
