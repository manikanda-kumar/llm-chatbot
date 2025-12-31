package routes

import (
	"banking-core/internal/handlers"
	"banking-core/internal/middleware"
	"banking-core/internal/repository"

	"github.com/gin-gonic/gin"
)

func Setup(router *gin.Engine, db *repository.DB) {
	// Middleware
	router.Use(middleware.CORS())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Initialize repositories
	customerRepo := repository.NewCustomerRepository(db)
	accountRepo := repository.NewAccountRepository(db)
	txnRepo := repository.NewTransactionRepository(db, accountRepo)
	transferRepo := repository.NewTransferRepository(db, accountRepo)
	cardRepo := repository.NewCardRepository(db)
	loanRepo := repository.NewLoanRepository(db)

	// Initialize handlers
	customerHandler := handlers.NewCustomerHandler(customerRepo, accountRepo)
	accountHandler := handlers.NewAccountHandler(accountRepo, txnRepo)
	txnHandler := handlers.NewTransactionHandler(txnRepo)
	transferHandler := handlers.NewTransferHandler(transferRepo)
	cardHandler := handlers.NewCardHandler(cardRepo, customerRepo)
	loanHandler := handlers.NewLoanHandler(loanRepo, customerRepo)

	// API v1 routes
	v1 := router.Group("/api/v1")
	v1.Use(middleware.APIKeyAuth())
	{
		// Customers
		customers := v1.Group("/customers")
		{
			customers.GET("", customerHandler.List)
			customers.POST("", customerHandler.Create)
			customers.POST("/search", customerHandler.Search)
			customers.GET("/search/name", customerHandler.SearchByName)
			customers.GET("/search/dob", customerHandler.SearchByDOB)
			customers.GET("/verify", customerHandler.VerifyIdentity)
			customers.GET("/status/:status", customerHandler.GetByStatus)
			customers.GET("/:id", customerHandler.GetByID)
			customers.GET("/:id/accounts", customerHandler.GetAccounts)
			customers.GET("/:id/cards", cardHandler.GetByCustomerID)
			customers.GET("/:id/loans", loanHandler.GetByCustomerID)
		}

		// Accounts
		accounts := v1.Group("/accounts")
		{
			accounts.GET("/:id", accountHandler.GetByID)
			accounts.GET("/:id/balance", accountHandler.GetBalance)
			accounts.GET("/:id/transactions", accountHandler.GetTransactions)
			accounts.POST("", accountHandler.Create)
		}

		// Transactions
		transactions := v1.Group("/transactions")
		{
			transactions.GET("/:id", txnHandler.GetByID)
			transactions.POST("", txnHandler.Create)
		}

		// Transfers
		transfers := v1.Group("/transfers")
		{
			transfers.GET("/:id", transferHandler.GetByID)
			transfers.POST("", transferHandler.Create)
		}

		// Cards
		cards := v1.Group("/cards")
		{
			cards.GET("/:id", cardHandler.GetByID)
			cards.POST("", cardHandler.Create)
			cards.POST("/:id/block", cardHandler.Block)
		}

		// Loans
		loans := v1.Group("/loans")
		{
			loans.GET("/:id", loanHandler.GetByID)
			loans.GET("/:id/schedule", loanHandler.GetPaymentSchedule)
			loans.POST("", loanHandler.Create)
		}
	}
}
