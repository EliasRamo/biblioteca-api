
package service

import (
	"errors"
	"time"

	"biblioteca-api/internal/models"
	"biblioteca-api/internal/repository"
)

const laptopLoanDurationDays = 14

type LaptopService struct {
	laptopRepo *repository.LaptopRepository
}

func NewLaptopService(laptopRepo *repository.LaptopRepository) *LaptopService {
	return &LaptopService{laptopRepo: laptopRepo}
}

// Create registra una nueva laptop.
func (s *LaptopService) Create(req models.LaptopRequest) (*models.Laptop, error) {
	laptop := &models.Laptop{
		SerialNumber: req.SerialNumber,
		Brand:        req.Brand,
		Model:        req.Model,
		Description:  req.Description,
	}

	if err := s.laptopRepo.Create(laptop); err != nil {
		return nil, err
	}

	return laptop, nil
}

// ListAvailable muestra las laptops disponibles.
func (s *LaptopService) ListAvailable() ([]models.Laptop, error) {
	return s.laptopRepo.FindAllAvailable()
}

// ListAll muestra todo el inventario.
func (s *LaptopService) ListAll() ([]models.Laptop, error) {
	return s.laptopRepo.FindAll()
}

// Borrow solicita una laptop por 14 días.
func (s *LaptopService) Borrow(userID, laptopID int) (*models.LaptopLoan, error) {
	laptop, err := s.laptopRepo.FindByID(laptopID)
	if err != nil {
		return nil, err
	}

	if laptop.Status != "available" {
		return nil, errors.New("la laptop no está disponible")
	}

	dueAt := time.Now().AddDate(0, 0, laptopLoanDurationDays)

	loan, err := s.laptopRepo.CreateLoan(userID, laptopID, dueAt)
	if err != nil {
		return nil, err
	}

	if loan == nil {
		return nil, errors.New("la laptop ya no está disponible")
	}

	return loan, nil
}

// MyLoans muestra los préstamos de un estudiante.
func (s *LaptopService) MyLoans(userID int) ([]models.LaptopLoan, error) {
	return s.laptopRepo.FindLoansByUser(userID)
}

// AllLoans muestra todos los préstamos.
func (s *LaptopService) AllLoans() ([]models.LaptopLoan, error) {
	return s.laptopRepo.FindAllLoans()
}

// Return registra la devolución de una laptop.
func (s *LaptopService) Return(userID, loanID int, condition string) error {
	if condition != "good" && condition != "damaged" {
		return errors.New("la condición debe ser good o damaged")
	}

	return s.laptopRepo.MarkReturned(loanID, userID, condition)
}
