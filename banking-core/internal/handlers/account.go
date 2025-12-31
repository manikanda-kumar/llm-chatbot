package handlers

import (
	"net/http"
	"strconv"

	"banking-core/internal/models"
	"banking-core/internal/repository"

	"github.com/gin-gonic/gin"
)

type AccountHandler struct {
	repo    *repository.AccountRepository
	txnRepo *repository.TransactionRepository
}

func NewAccountHandler(repo *repository.AccountRepository, txnRepo *repository.TransactionRepository) *AccountHandler {
	return &AccountHandler{repo: repo, txnRepo: txnRepo}
}

func (h *AccountHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	account, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	c.JSON(http.StatusOK, account)
}

func (h *AccountHandler) GetBalance(c *gin.Context) {
	id := c.Param("id")
	balance, err := h.repo.GetBalance(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}
	c.JSON(http.StatusOK, balance)
}

func (h *AccountHandler) GetTransactions(c *gin.Context) {
	id := c.Param("id")

	// Verify account exists
	_, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	transactions, err := h.txnRepo.GetByAccountID(id, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, transactions)
}

func (h *AccountHandler) Create(c *gin.Context) {
	var req models.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account, err := h.repo.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, account)
}
