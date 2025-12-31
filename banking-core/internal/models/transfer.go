package models

import "time"

type TransferStatus string

const (
	TransferStatusPending   TransferStatus = "pending"
	TransferStatusCompleted TransferStatus = "completed"
	TransferStatusFailed    TransferStatus = "failed"
)

type Transfer struct {
	ID            string         `db:"id" json:"id"`
	FromAccountID string         `db:"from_account_id" json:"from_account_id"`
	ToAccountID   string         `db:"to_account_id" json:"to_account_id"`
	Amount        float64        `db:"amount" json:"amount"`
	Status        TransferStatus `db:"status" json:"status"`
	CreatedAt     time.Time      `db:"created_at" json:"created_at"`
}

type CreateTransferRequest struct {
	FromAccountID string  `json:"from_account_id" binding:"required"`
	ToAccountID   string  `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
}
