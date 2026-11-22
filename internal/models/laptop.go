package models

import "time"

type Laptop struct {
	ID           int       `json:"id" db:"id"`
	SerialNumber string    `json:"serial_number" db:"serial_number"`
	Brand        string    `json:"brand" db:"brand"`
	Model        string    `json:"model" db:"model"`
	Description  string    `json:"description" db:"description"`
	Status       string    `json:"status" db:"status"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

type LaptopRequest struct {
	SerialNumber string `json:"serial_number" binding:"required"`
	Brand        string `json:"brand" binding:"required"`
	Model        string `json:"model" binding:"required"`
	Description  string `json:"description"`
}

type LaptopLoan struct {
	ID              int        `json:"id" db:"id"`
	UserID          int        `json:"user_id" db:"user_id"`
	LaptopID        int        `json:"laptop_id" db:"laptop_id"`
	LoanedAt        time.Time  `json:"loaned_at" db:"loaned_at"`
	DueAt           time.Time  `json:"due_at" db:"due_at"`
	ReturnedAt      *time.Time `json:"returned_at" db:"returned_at"`
	Status          string     `json:"status" db:"status"`
	ReturnCondition *string    `json:"return_condition" db:"return_condition"`
	ActiveLaptopID  *int       `json:"-" db:"active_laptop_id"`
}

type LaptopLoanRequest struct {
	LaptopID int `json:"laptop_id" binding:"required"`
}