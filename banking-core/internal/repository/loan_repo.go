package repository

import (
	"math"

	"banking-core/internal/models"

	"github.com/google/uuid"
)

type LoanRepository struct {
	db *DB
}

func NewLoanRepository(db *DB) *LoanRepository {
	return &LoanRepository{db: db}
}

func (r *LoanRepository) GetByID(id string) (*models.Loan, error) {
	var loan models.Loan
	err := r.db.Get(&loan, "SELECT * FROM loans WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &loan, nil
}

func (r *LoanRepository) GetByCustomerID(customerID string) ([]models.Loan, error) {
	var loans []models.Loan
	err := r.db.Select(&loans, "SELECT * FROM loans WHERE customer_id = ? ORDER BY created_at DESC", customerID)
	return loans, err
}

func (r *LoanRepository) Create(req *models.CreateLoanRequest) (*models.Loan, error) {
	id := uuid.New().String()

	// Calculate monthly payment using amortization formula
	monthlyRate := req.InterestRate / 100 / 12
	monthlyPayment := req.Principal * (monthlyRate * math.Pow(1+monthlyRate, float64(req.TermMonths))) /
		(math.Pow(1+monthlyRate, float64(req.TermMonths)) - 1)

	_, err := r.db.Exec(
		"INSERT INTO loans (id, customer_id, loan_type, principal, interest_rate, term_months, monthly_payment, balance, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'active')",
		id, req.CustomerID, req.LoanType, req.Principal, req.InterestRate, req.TermMonths, monthlyPayment, req.Principal,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *LoanRepository) GetPaymentSchedule(id string) ([]models.PaymentScheduleItem, error) {
	loan, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	schedule := make([]models.PaymentScheduleItem, 0, loan.TermMonths)
	balance := loan.Principal
	monthlyRate := loan.InterestRate / 100 / 12

	for month := 1; month <= loan.TermMonths && balance > 0; month++ {
		interest := balance * monthlyRate
		principal := loan.MonthlyPayment - interest
		if principal > balance {
			principal = balance
		}
		balance -= principal

		schedule = append(schedule, models.PaymentScheduleItem{
			Month:     month,
			Payment:   loan.MonthlyPayment,
			Principal: math.Round(principal*100) / 100,
			Interest:  math.Round(interest*100) / 100,
			Balance:   math.Round(balance*100) / 100,
		})
	}

	return schedule, nil
}
