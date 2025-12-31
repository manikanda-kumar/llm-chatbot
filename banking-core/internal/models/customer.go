package models

import "time"

type Gender string

const (
	GenderMale   Gender = "male"
	GenderFemale Gender = "female"
	GenderOther  Gender = "other"
)

type CustomerStatus string

const (
	CustomerStatusActive   CustomerStatus = "active"
	CustomerStatusInactive CustomerStatus = "inactive"
	CustomerStatusSuspended CustomerStatus = "suspended"
)

type Customer struct {
	ID          string         `db:"id" json:"id"`
	FirstName   string         `db:"first_name" json:"first_name"`
	LastName    string         `db:"last_name" json:"last_name"`
	Email       string         `db:"email" json:"email"`
	Phone       string         `db:"phone" json:"phone,omitempty"`
	DateOfBirth string         `db:"date_of_birth" json:"date_of_birth"` // YYYY-MM-DD
	Gender      Gender         `db:"gender" json:"gender"`
	SIN         string         `db:"sin" json:"sin"` // Social Insurance Number (masked)
	Address     string         `db:"address" json:"address,omitempty"`
	City        string         `db:"city" json:"city,omitempty"`
	Province    string         `db:"province" json:"province,omitempty"`
	PostalCode  string         `db:"postal_code" json:"postal_code,omitempty"`
	Status      CustomerStatus `db:"status" json:"status"`
	CreatedAt   time.Time      `db:"created_at" json:"created_at"`
}

type CreateCustomerRequest struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Phone       string `json:"phone"`
	DateOfBirth string `json:"date_of_birth" binding:"required"`
	Gender      Gender `json:"gender" binding:"required"`
	SIN         string `json:"sin"`
	Address     string `json:"address"`
	City        string `json:"city"`
	Province    string `json:"province"`
	PostalCode  string `json:"postal_code"`
}

// CustomerSearchRequest for searching customers with multiple criteria
type CustomerSearchRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	DateOfBirth string `json:"date_of_birth"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	SIN         string `json:"sin"`
}
