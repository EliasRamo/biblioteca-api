
package repository

import (
	"time"

	"github.com/jmoiron/sqlx"

	"biblioteca-api/internal/models"
)

type LaptopRepository struct {
	db *sqlx.DB
}

func NewLaptopRepository(db *sqlx.DB) *LaptopRepository {
	return &LaptopRepository{db: db}
}

// Create registra una laptop.
func (r *LaptopRepository) Create(l *models.Laptop) error {
	query := `INSERT INTO laptops (serial_number, brand, model, description, status)
	          VALUES (?, ?, ?, ?, 'available')`

	result, err := r.db.Exec(query, l.SerialNumber, l.Brand, l.Model, l.Description)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}

	l.ID = int(id)
	l.Status = "available"

	return nil
}

// FindAllAvailable devuelve las laptops disponibles.
func (r *LaptopRepository) FindAllAvailable() ([]models.Laptop, error) {
	var laptops []models.Laptop

	query := `SELECT * FROM laptops WHERE status = 'available' ORDER BY id DESC`

	if err := r.db.Select(&laptops, query); err != nil {
		return nil, err
	}

	return laptops, nil
}

// FindAll devuelve todas las laptops.
func (r *LaptopRepository) FindAll() ([]models.Laptop, error) {
	var laptops []models.Laptop

	query := `SELECT * FROM laptops ORDER BY id DESC`

	if err := r.db.Select(&laptops, query); err != nil {
		return nil, err
	}

	return laptops, nil
}

// FindByID busca una laptop por su identificador.
func (r *LaptopRepository) FindByID(id int) (*models.Laptop, error) {
	var laptop models.Laptop

	query := `SELECT * FROM laptops WHERE id = ?`

	if err := r.db.Get(&laptop, query, id); err != nil {
		return nil, err
	}

	return &laptop, nil
}

// CreateLoan registra un préstamo y cambia el estado de la laptop.
func (r *LaptopRepository) CreateLoan(userID, laptopID int, dueAt time.Time) (*models.LaptopLoan, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result, err := tx.Exec(
		`UPDATE laptops SET status = 'loaned' WHERE id = ? AND status = 'available'`,
		laptopID,
	)
	if err != nil {
		return nil, err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return nil, err
	}

	if rows == 0 {
		return nil, nil
	}

	result, err = tx.Exec(
		`INSERT INTO laptop_loans (user_id, laptop_id, due_at, status)
		 VALUES (?, ?, ?, 'active')`,
		userID, laptopID, dueAt,
	)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.FindLoanByID(int(id))
}

// FindLoanByID busca un préstamo por su identificador.
func (r *LaptopRepository) FindLoanByID(id int) (*models.LaptopLoan, error) {
	var loan models.LaptopLoan

	query := `SELECT * FROM laptop_loans WHERE id = ?`

	if err := r.db.Get(&loan, query, id); err != nil {
		return nil, err
	}

	return &loan, nil
}

// FindLoansByUser devuelve los préstamos de un estudiante.
func (r *LaptopRepository) FindLoansByUser(userID int) ([]models.LaptopLoan, error) {
	var loans []models.LaptopLoan

	query := `SELECT * FROM laptop_loans WHERE user_id = ? ORDER BY loaned_at DESC`

	if err := r.db.Select(&loans, query, userID); err != nil {
		return nil, err
	}

	return loans, nil
}

// FindAllLoans devuelve todos los préstamos de laptops.
func (r *LaptopRepository) FindAllLoans() ([]models.LaptopLoan, error) {
	var loans []models.LaptopLoan

	query := `SELECT * FROM laptop_loans ORDER BY loaned_at DESC`

	if err := r.db.Select(&loans, query); err != nil {
		return nil, err
	}

	return loans, nil
}

// MarkReturned registra la devolución y actualiza el estado de la laptop.
func (r *LaptopRepository) MarkReturned(id, userID int, condition string) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var laptopID int

	err = tx.Get(
		&laptopID,
		`SELECT laptop_id FROM laptop_loans
		 WHERE id = ? AND user_id = ? AND status = 'active'`,
		id, userID,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE laptop_loans
		 SET status = 'returned', returned_at = NOW(), return_condition = ?
		 WHERE id = ? AND user_id = ? AND status = 'active'`,
		condition, id, userID,
	)
	if err != nil {
		return err
	}

	newStatus := "maintenance"
	if condition == "good" {
		newStatus = "available"
	}

	_, err = tx.Exec(
		`UPDATE laptops SET status = ? WHERE id = ?`,
		newStatus, laptopID,
	)
	if err != nil {
		return err
	}

	return tx.Commit()
}
