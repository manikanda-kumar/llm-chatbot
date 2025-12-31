package repository

import (
	"crypto/rand"
	"fmt"
	"time"

	"banking-core/internal/models"

	"github.com/google/uuid"
)

type CardRepository struct {
	db *DB
}

func NewCardRepository(db *DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) GetByID(id string) (*models.Card, error) {
	var card models.Card
	err := r.db.Get(&card, "SELECT * FROM cards WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) GetByCustomerID(customerID string) ([]models.Card, error) {
	var cards []models.Card
	err := r.db.Select(&cards, "SELECT * FROM cards WHERE customer_id = ? ORDER BY created_at DESC", customerID)
	return cards, err
}

func (r *CardRepository) Create(req *models.CreateCardRequest) (*models.Card, error) {
	id := uuid.New().String()

	// Generate masked card number
	cardNumber := generateMaskedCardNumber()

	// Generate expiry date (3 years from now)
	expiryDate := time.Now().AddDate(3, 0, 0).Format("01/06")

	_, err := r.db.Exec(
		"INSERT INTO cards (id, customer_id, account_id, card_number, card_type, expiry_date, status) VALUES (?, ?, ?, ?, ?, ?, 'active')",
		id, req.CustomerID, req.AccountID, cardNumber, req.CardType, expiryDate,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *CardRepository) Block(id string) error {
	_, err := r.db.Exec("UPDATE cards SET status = 'blocked' WHERE id = ?", id)
	return err
}

func (r *CardRepository) Unblock(id string) error {
	_, err := r.db.Exec("UPDATE cards SET status = 'active' WHERE id = ?", id)
	return err
}

func generateMaskedCardNumber() string {
	b := make([]byte, 2)
	rand.Read(b)
	lastFour := fmt.Sprintf("%04d", int(b[0])<<8|int(b[1])%10000)
	return "****-****-****-" + lastFour
}
