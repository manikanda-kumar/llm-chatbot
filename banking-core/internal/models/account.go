package models

import "time"

type AccountType string

const (
	AccountTypeChecking AccountType = "checking"
	AccountTypeSavings  AccountType = "savings"
	AccountTypeCredit   AccountType = "credit"
)

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
	AccountStatusFrozen   AccountStatus = "frozen"
)

type Account struct {
	ID          string        `db:"id" json:"id"`
	CustomerID  string        `db:"customer_id" json:"customer_id"`
	AccountType AccountType   `db:"account_type" json:"account_type"`
	Balance     float64       `db:"balance" json:"balance"`
	Currency    string        `db:"currency" json:"currency"`
	Status      AccountStatus `db:"status" json:"status"`
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
}

type CreateAccountRequest struct {
	CustomerID  string      `json:"customer_id" binding:"required"`
	AccountType AccountType `json:"account_type" binding:"required"`
	Currency    string      `json:"currency"`
}

type BalanceResponse struct {
	AccountID string  `json:"account_id"`
	Balance   float64 `json:"balance"`
	Currency  string  `json:"currency"`
}
