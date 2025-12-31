package handlers

import (
	"net/http"

	"banking-core/internal/models"
	"banking-core/internal/repository"

	"github.com/gin-gonic/gin"
)

type CardHandler struct {
	repo         *repository.CardRepository
	customerRepo *repository.CustomerRepository
}

func NewCardHandler(repo *repository.CardRepository, customerRepo *repository.CustomerRepository) *CardHandler {
	return &CardHandler{repo: repo, customerRepo: customerRepo}
}

func (h *CardHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	card, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}
	c.JSON(http.StatusOK, card)
}

func (h *CardHandler) GetByCustomerID(c *gin.Context) {
	customerID := c.Param("id")

	// Verify customer exists
	_, err := h.customerRepo.GetByID(customerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	cards, err := h.repo.GetByCustomerID(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cards)
}

func (h *CardHandler) Create(c *gin.Context) {
	var req models.CreateCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	card, err := h.repo.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, card)
}

func (h *CardHandler) Block(c *gin.Context) {
	id := c.Param("id")

	// Verify card exists
	_, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Card not found"})
		return
	}

	if err := h.repo.Block(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Card blocked successfully"})
}
