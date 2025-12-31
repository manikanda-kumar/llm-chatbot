package handlers

import (
	"net/http"

	"banking-core/internal/models"
	"banking-core/internal/repository"

	"github.com/gin-gonic/gin"
)

type TransferHandler struct {
	repo *repository.TransferRepository
}

func NewTransferHandler(repo *repository.TransferRepository) *TransferHandler {
	return &TransferHandler{repo: repo}
}

func (h *TransferHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	transfer, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transfer not found"})
		return
	}
	c.JSON(http.StatusOK, transfer)
}

func (h *TransferHandler) Create(c *gin.Context) {
	var req models.CreateTransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	transfer, err := h.repo.Create(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, transfer)
}
