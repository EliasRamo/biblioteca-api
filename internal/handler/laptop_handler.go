package handler

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"

	"biblioteca-api/internal/models"
	"biblioteca-api/internal/service"
)

type LaptopHandler struct {
	laptopService *service.LaptopService
}

func NewLaptopHandler(laptopService *service.LaptopService) *LaptopHandler {
	return &LaptopHandler{laptopService: laptopService}
}

// Create registra una laptop.
func (h *LaptopHandler) Create(c *gin.Context) {
	var req models.LaptopRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	laptop, err := h.laptopService.Create(req)
	if err != nil {
		var mysqlErr *mysql.MySQLError

		if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "ya existe una laptop con ese número de serie",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo registrar la laptop"})
		return
	}

	c.JSON(http.StatusCreated, laptop)
}

// ListAvailable muestra las laptops disponibles.
func (h *LaptopHandler) ListAvailable(c *gin.Context) {
	laptops, err := h.laptopService.ListAvailable()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron obtener las laptops"})
		return
	}

	c.JSON(http.StatusOK, laptops)
}

// ListAll muestra todo el inventario.
func (h *LaptopHandler) ListAll(c *gin.Context) {
	laptops, err := h.laptopService.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudo obtener el inventario"})
		return
	}

	c.JSON(http.StatusOK, laptops)
}

// Borrow solicita una laptop.
func (h *LaptopHandler) Borrow(c *gin.Context) {
	userID := c.GetInt("user_id")

	var req models.LaptopLoanRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan, err := h.laptopService.Borrow(userID, req.LaptopID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "laptop no encontrada"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, loan)
}

// MyLoans muestra los préstamos del usuario autenticado.
func (h *LaptopHandler) MyLoans(c *gin.Context) {
	userID := c.GetInt("user_id")

	loans, err := h.laptopService.MyLoans(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron obtener los préstamos"})
		return
	}

	c.JSON(http.StatusOK, loans)
}

// AllLoans muestra todos los préstamos.
func (h *LaptopHandler) AllLoans(c *gin.Context) {
	loans, err := h.laptopService.AllLoans()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "no se pudieron obtener los préstamos"})
		return
	}

	c.JSON(http.StatusOK, loans)
}

// Return registra la devolución de una laptop.
func (h *LaptopHandler) Return(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	var req struct {
		Condition string `json:"condition" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.laptopService.Return(c.GetInt("user_id"), id, req.Condition); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "préstamo activo no encontrado"})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "laptop devuelta correctamente"})
}
