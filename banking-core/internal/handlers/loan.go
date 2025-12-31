package handlers

import (
	"net/http"

	"banking-core/internal/models"
	"banking-core/internal/repository"

	"github.com/gin-gonic/gin"
)

type LoanHandler struct {
	repo         *repository.LoanRepository
	customerRepo *repository.CustomerRepository
}

func NewLoanHandler(repo *repository.LoanRepository, customerRepo *repository.CustomerRepository) *LoanHandler {
	return &LoanHandler{repo: repo, customerRepo: customerRepo}
}

func (h *LoanHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	loan, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}
	c.JSON(http.StatusOK, loan)
}

func (h *LoanHandler) GetByCustomerID(c *gin.Context) {
	customerID := c.Param("id")

	// Verify customer exists
	_, err := h.customerRepo.GetByID(customerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	loans, err := h.repo.GetByCustomerID(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, loans)
}

func (h *LoanHandler) GetPaymentSchedule(c *gin.Context) {
	id := c.Param("id")
	schedule, err := h.repo.GetPaymentSchedule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}
	c.JSON(http.StatusOK, schedule)
}

func (h *LoanHandler) Create(c *gin.Context) {
	var req models.CreateLoanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	loan, err := h.repo.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, loan)
}
