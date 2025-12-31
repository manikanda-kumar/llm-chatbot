package repository

import (
	"errors"

	"banking-core/internal/models"

	"github.com/google/uuid"
)

type TransferRepository struct {
	db          *DB
	accountRepo *AccountRepository
}

func NewTransferRepository(db *DB, accountRepo *AccountRepository) *TransferRepository {
	return &TransferRepository{db: db, accountRepo: accountRepo}
}

func (r *TransferRepository) GetByID(id string) (*models.Transfer, error) {
	var transfer models.Transfer
	err := r.db.Get(&transfer, "SELECT * FROM transfers WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &transfer, nil
}

func (r *TransferRepository) Create(req *models.CreateTransferRequest) (*models.Transfer, error) {
	// Validate accounts exist
	fromAccount, err := r.accountRepo.GetByID(req.FromAccountID)
	if err != nil {
		return nil, errors.New("source account not found")
	}

	_, err = r.accountRepo.GetByID(req.ToAccountID)
	if err != nil {
		return nil, errors.New("destination account not found")
	}

	// Check sufficient funds
	if fromAccount.Balance < req.Amount {
		return nil, errors.New("insufficient funds")
	}

	// Start transaction
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	// Create transfer record
	transferID := uuid.New().String()
	_, err = tx.Exec(
		"INSERT INTO transfers (id, from_account_id, to_account_id, amount, status) VALUES (?, ?, ?, ?, 'completed')",
		transferID, req.FromAccountID, req.ToAccountID, req.Amount,
	)
	if err != nil {
		return nil, err
	}

	// Debit from account
	txnOutID := uuid.New().String()
	_, err = tx.Exec(
		"INSERT INTO transactions (id, account_id, type, amount, description) VALUES (?, ?, 'transfer_out', ?, ?)",
		txnOutID, req.FromAccountID, req.Amount, "Transfer to "+req.ToAccountID,
	)
	if err != nil {
		return nil, err
	}

	// Credit to account
	txnInID := uuid.New().String()
	_, err = tx.Exec(
		"INSERT INTO transactions (id, account_id, type, amount, description) VALUES (?, ?, 'transfer_in', ?, ?)",
		txnInID, req.ToAccountID, req.Amount, "Transfer from "+req.FromAccountID,
	)
	if err != nil {
		return nil, err
	}

	// Update balances
	_, err = tx.Exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", req.Amount, req.FromAccountID)
	if err != nil {
		return nil, err
	}
	_, err = tx.Exec("UPDATE accounts SET balance = balance + ? WHERE id = ?", req.Amount, req.ToAccountID)
	if err != nil {
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetByID(transferID)
}
