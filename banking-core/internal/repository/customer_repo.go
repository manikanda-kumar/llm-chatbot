package repository

import (
	"fmt"
	"strings"

	"banking-core/internal/models"

	"github.com/google/uuid"
)

type CustomerRepository struct {
	db *DB
}

func NewCustomerRepository(db *DB) *CustomerRepository {
	return &CustomerRepository{db: db}
}

func (r *CustomerRepository) GetByID(id string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Get(&customer, "SELECT * FROM customers WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) GetByEmail(email string) (*models.Customer, error) {
	var customer models.Customer
	err := r.db.Get(&customer, "SELECT * FROM customers WHERE email = ?", email)
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *CustomerRepository) Create(req *models.CreateCustomerRequest) (*models.Customer, error) {
	id := uuid.New().String()
	_, err := r.db.Exec(
		`INSERT INTO customers (id, first_name, last_name, email, phone, date_of_birth, gender, sin, address, city, province, postal_code, status)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 'active')`,
		id, req.FirstName, req.LastName, req.Email, req.Phone,
		req.DateOfBirth, req.Gender, req.SIN,
		req.Address, req.City, req.Province, req.PostalCode,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(id)
}

func (r *CustomerRepository) List(limit, offset int) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Select(&customers, "SELECT * FROM customers ORDER BY created_at DESC LIMIT ? OFFSET ?", limit, offset)
	return customers, err
}

// Search finds customers matching the given criteria
// Returns multiple matches to enforce verification before returning data
func (r *CustomerRepository) Search(req *models.CustomerSearchRequest) ([]models.Customer, error) {
	var customers []models.Customer
	var conditions []string
	var args []interface{}

	if req.FirstName != "" {
		conditions = append(conditions, "LOWER(first_name) = LOWER(?)")
		args = append(args, req.FirstName)
	}
	if req.LastName != "" {
		conditions = append(conditions, "LOWER(last_name) = LOWER(?)")
		args = append(args, req.LastName)
	}
	if req.DateOfBirth != "" {
		conditions = append(conditions, "date_of_birth = ?")
		args = append(args, req.DateOfBirth)
	}
	if req.Email != "" {
		conditions = append(conditions, "LOWER(email) = LOWER(?)")
		args = append(args, req.Email)
	}
	if req.Phone != "" {
		conditions = append(conditions, "phone = ?")
		args = append(args, req.Phone)
	}
	if req.SIN != "" {
		conditions = append(conditions, "sin = ?")
		args = append(args, req.SIN)
	}

	if len(conditions) == 0 {
		return nil, fmt.Errorf("at least one search criteria required")
	}

	query := "SELECT * FROM customers WHERE " + strings.Join(conditions, " AND ") + " ORDER BY last_name, first_name"
	err := r.db.Select(&customers, query, args...)
	return customers, err
}

// GetByNameAndDOB finds customers by name and DOB - critical for identity verification
func (r *CustomerRepository) GetByNameAndDOB(firstName, lastName, dob string) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Select(&customers,
		"SELECT * FROM customers WHERE LOWER(first_name) = LOWER(?) AND LOWER(last_name) = LOWER(?) AND date_of_birth = ?",
		firstName, lastName, dob)
	return customers, err
}

// GetByName finds all customers with matching name - shows potential duplicates
func (r *CustomerRepository) GetByName(firstName, lastName string) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Select(&customers,
		"SELECT * FROM customers WHERE LOWER(first_name) = LOWER(?) AND LOWER(last_name) = LOWER(?)",
		firstName, lastName)
	return customers, err
}

// GetByDOB finds all customers with the same DOB
func (r *CustomerRepository) GetByDOB(dob string) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Select(&customers,
		"SELECT * FROM customers WHERE date_of_birth = ?",
		dob)
	return customers, err
}

// GetActiveCustomers returns only active customers
func (r *CustomerRepository) GetActiveCustomers(limit, offset int) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Select(&customers,
		"SELECT * FROM customers WHERE status = 'active' ORDER BY last_name, first_name LIMIT ? OFFSET ?",
		limit, offset)
	return customers, err
}

// GetByStatus returns customers with a specific status
func (r *CustomerRepository) GetByStatus(status string) ([]models.Customer, error) {
	var customers []models.Customer
	err := r.db.Select(&customers,
		"SELECT * FROM customers WHERE status = ? ORDER BY last_name, first_name",
		status)
	return customers, err
}

// UpdateStatus changes customer status
func (r *CustomerRepository) UpdateStatus(id string, status models.CustomerStatus) error {
	_, err := r.db.Exec("UPDATE customers SET status = ? WHERE id = ?", status, id)
	return err
}
