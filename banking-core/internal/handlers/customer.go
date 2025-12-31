package handlers

import (
	"net/http"

	"banking-core/internal/models"
	"banking-core/internal/repository"

	"github.com/gin-gonic/gin"
)

type CustomerHandler struct {
	repo        *repository.CustomerRepository
	accountRepo *repository.AccountRepository
}

func NewCustomerHandler(repo *repository.CustomerRepository, accountRepo *repository.AccountRepository) *CustomerHandler {
	return &CustomerHandler{repo: repo, accountRepo: accountRepo}
}

func (h *CustomerHandler) GetByID(c *gin.Context) {
	id := c.Param("id")
	customer, err := h.repo.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) GetAccounts(c *gin.Context) {
	customerID := c.Param("id")

	// Verify customer exists and is active
	customer, err := h.repo.GetByID(customerID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Customer not found"})
		return
	}

	// Check customer status
	if customer.Status != models.CustomerStatusActive {
		c.JSON(http.StatusForbidden, gin.H{
			"error":  "Customer account is not active",
			"status": customer.Status,
		})
		return
	}

	accounts, err := h.accountRepo.GetByCustomerID(customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, accounts)
}

func (h *CustomerHandler) Create(c *gin.Context) {
	var req models.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err := h.repo.Create(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, customer)
}

func (h *CustomerHandler) List(c *gin.Context) {
	customers, err := h.repo.List(100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customers)
}

// Search finds customers based on criteria - returns multiple matches for verification
func (h *CustomerHandler) Search(c *gin.Context) {
	var req models.CustomerSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customers, err := h.repo.Search(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Return with warning if multiple matches found
	response := gin.H{
		"customers":     customers,
		"match_count":   len(customers),
		"requires_verification": len(customers) > 1,
	}

	if len(customers) > 1 {
		response["warning"] = "Multiple customers match the criteria. Additional verification required."
	}

	c.JSON(http.StatusOK, response)
}

// SearchByName finds all customers with matching name - shows duplicates
func (h *CustomerHandler) SearchByName(c *gin.Context) {
	firstName := c.Query("first_name")
	lastName := c.Query("last_name")

	if firstName == "" || lastName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Both first_name and last_name required"})
		return
	}

	customers, err := h.repo.GetByName(firstName, lastName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If multiple matches, require additional verification
	response := gin.H{
		"customers":   customers,
		"match_count": len(customers),
	}

	if len(customers) > 1 {
		response["warning"] = "Multiple customers with same name. Verify DOB, gender, or SIN to confirm identity."
		response["requires_verification"] = true
	}

	c.JSON(http.StatusOK, response)
}

// SearchByDOB finds all customers born on a specific date
func (h *CustomerHandler) SearchByDOB(c *gin.Context) {
	dob := c.Query("dob")
	if dob == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dob parameter required (YYYY-MM-DD)"})
		return
	}

	customers, err := h.repo.GetByDOB(dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers":   customers,
		"match_count": len(customers),
		"dob":         dob,
	})
}

// VerifyIdentity performs strict identity verification
func (h *CustomerHandler) VerifyIdentity(c *gin.Context) {
	firstName := c.Query("first_name")
	lastName := c.Query("last_name")
	dob := c.Query("dob")

	if firstName == "" || lastName == "" || dob == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "first_name, last_name, and dob are all required"})
		return
	}

	customers, err := h.repo.GetByNameAndDOB(firstName, lastName, dob)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	response := gin.H{
		"customers":   customers,
		"match_count": len(customers),
	}

	if len(customers) == 0 {
		response["verified"] = false
		response["message"] = "No customer found with the provided details"
	} else if len(customers) == 1 {
		response["verified"] = true
		response["customer_id"] = customers[0].ID
	} else {
		response["verified"] = false
		response["warning"] = "Multiple customers match name and DOB. Additional verification required (gender or SIN)."
		response["requires_additional_verification"] = true
	}

	c.JSON(http.StatusOK, response)
}

// GetByStatus returns customers with a specific status
func (h *CustomerHandler) GetByStatus(c *gin.Context) {
	status := c.Param("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status required"})
		return
	}

	customers, err := h.repo.GetByStatus(status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"customers": customers,
		"count":     len(customers),
	})
}
