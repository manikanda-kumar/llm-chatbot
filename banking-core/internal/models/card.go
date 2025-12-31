package models

import "time"

type CardType string

const (
	CardTypeDebit  CardType = "debit"
	CardTypeCredit CardType = "credit"
)

type CardStatus string

const (
	CardStatusActive  CardStatus = "active"
	CardStatusBlocked CardStatus = "blocked"
	CardStatusExpired CardStatus = "expired"
)

type Card struct {
	ID         string     `db:"id" json:"id"`
	CustomerID string     `db:"customer_id" json:"customer_id"`
	AccountID  string     `db:"account_id" json:"account_id"`
	CardNumber string     `db:"card_number" json:"card_number"` // Masked
	CardType   CardType   `db:"card_type" json:"card_type"`
	ExpiryDate string     `db:"expiry_date" json:"expiry_date"`
	Status     CardStatus `db:"status" json:"status"`
	CreatedAt  time.Time  `db:"created_at" json:"created_at"`
}

type CreateCardRequest struct {
	CustomerID string   `json:"customer_id" binding:"required"`
	AccountID  string   `json:"account_id" binding:"required"`
	CardType   CardType `json:"card_type" binding:"required"`
}
