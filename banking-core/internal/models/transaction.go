package models

import "time"

type TransactionType string

const (
	TransactionTypeDeposit     TransactionType = "deposit"
	TransactionTypeWithdrawal  TransactionType = "withdrawal"
	TransactionTypeTransferIn  TransactionType = "transfer_in"
	TransactionTypeTransferOut TransactionType = "transfer_out"
)

type Transaction struct {
	ID          string          `db:"id" json:"id"`
	AccountID   string          `db:"account_id" json:"account_id"`
	Type        TransactionType `db:"type" json:"type"`
	Amount      float64         `db:"amount" json:"amount"`
	Description string          `db:"description" json:"description,omitempty"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
}

type CreateTransactionRequest struct {
	AccountID   string          `json:"account_id" binding:"required"`
	Type        TransactionType `json:"type" binding:"required"`
	Amount      float64         `json:"amount" binding:"required,gt=0"`
	Description string          `json:"description"`
}
