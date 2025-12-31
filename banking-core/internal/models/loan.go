package models

import "time"

type LoanType string

const (
	LoanTypeMortgage LoanType = "mortgage"
	LoanTypePersonal LoanType = "personal"
	LoanTypeAuto     LoanType = "auto"
)

type LoanStatus string

const (
	LoanStatusActive    LoanStatus = "active"
	LoanStatusPaidOff   LoanStatus = "paid_off"
	LoanStatusDefaulted LoanStatus = "defaulted"
)

type Loan struct {
	ID             string     `db:"id" json:"id"`
	CustomerID     string     `db:"customer_id" json:"customer_id"`
	LoanType       LoanType   `db:"loan_type" json:"loan_type"`
	Principal      float64    `db:"principal" json:"principal"`
	InterestRate   float64    `db:"interest_rate" json:"interest_rate"`
	TermMonths     int        `db:"term_months" json:"term_months"`
	MonthlyPayment float64    `db:"monthly_payment" json:"monthly_payment"`
	Balance        float64    `db:"balance" json:"balance"`
	Status         LoanStatus `db:"status" json:"status"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
}

type PaymentScheduleItem struct {
	Month     int     `json:"month"`
	Payment   float64 `json:"payment"`
	Principal float64 `json:"principal"`
	Interest  float64 `json:"interest"`
	Balance   float64 `json:"balance"`
}

type CreateLoanRequest struct {
	CustomerID   string   `json:"customer_id" binding:"required"`
	LoanType     LoanType `json:"loan_type" binding:"required"`
	Principal    float64  `json:"principal" binding:"required,gt=0"`
	InterestRate float64  `json:"interest_rate" binding:"required,gt=0"`
	TermMonths   int      `json:"term_months" binding:"required,gt=0"`
}
