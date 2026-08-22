package handler

import (
	"Wallet-App/model"
	"Wallet-App/service"
	"errors"
	"Wallet-App/utils"
	"github.com/gin-gonic/gin"
)

type Transaction_handler struct {
	service service.Transaction_service
}

func Init_transaction_handler(service service.Transaction_service) *Transaction_handler {
	new_handler := Transaction_handler{service: service}
	return &new_handler
}

// @Summary	 		Get current user transactions.
// @Description 	Get all user transactions with optional filtering by category or date range (from & to).
// @Tags			Transactions
// @Param			category query 	string false		"Filter transactions by category"
// @Param			from query 	   	string false		"Start date for filter"
// @Param			to query 	   	string false		"End date for filter"
// @Produce			json
// @Success			200  {array}	model.Transaction	"List of transactions"
// @Failure			400	 {object}	map[string]string	"Error message"
// @Failure			404	 {object}	map[string]string	"Wallet not found"
// @Failure			500	 {object}	map[string]string	"Server error"
// @Router			/transactions [get]
func (handler *Transaction_handler) Get_transactions_handler(c *gin.Context) {
	//get the user claims.
	user, err := utils.Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//1. check if the category query param exists.
	//get the query param that define category.
	category, exists := c.GetQuery("category")
	if exists {
		//get all transactions of the target category.
		transactions, err := handler.service.Get_transactions_by_category(category, user)
		//check the result.
		switch {
		case err == nil:
			c.JSON(200, transactions)
		case errors.Is(err, service.ErrEmptyCategory):
			c.JSON(400, gin.H{"error": "Category cannot be empty"})
		case errors.Is(err, service.ErrWalletNotFound):
			c.JSON(404, gin.H{"error": "Wallet not found"})
		default:
			c.JSON(500, gin.H{"error": "Server error"})
		}
		return
	}
	//2. if category not found check the from and to params.
	//get the query params that define dates.
	start, exists := c.GetQuery("from")
	if exists {
		end, exists := c.GetQuery("to")
		if !exists {
			c.JSON(400, gin.H{"error": "Missing end date"})
			return
		}
		//get all transactions at that time.
		transactions, err := handler.service.Get_transactions_by_date(start, end, user)
		//check the result.
		switch {
		case err == nil:
			c.JSON(200, transactions)
		case errors.Is(err, service.ErrInvalidDate):
			c.JSON(400, gin.H{"error": "Invalid date format"})
		case errors.Is(err, service.ErrInvalidDateRange):
			c.JSON(400, gin.H{"error": "Start date cannot be after the end date"})
		case errors.Is(err, service.ErrWalletNotFound):
			c.JSON(404, gin.H{"error": "Wallet not found"})
		default:
			c.JSON(500, gin.H{"error": "Server error"})
		}
		return
	}
	//3. if there is no params, get all transactions of the user.
	transactions, err := handler.service.Get_user_transactions(user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(200, transactions)
	case errors.Is(err, service.ErrWalletNotFound):
		c.JSON(404, gin.H{"error": "Wallet not found"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}

// @Summary	 		Get current user transactions summary.
// @Description 	Get financial summary of transaction totals grouped by category for the current month.
// @Tags			Transactions
// @Produce			json
// @Success			200  {array}	model.Transaction_summary	"Transactions summary details"
// @Failure			404	 {object}	map[string]string			"Wallet not found"
// @Failure			500	 {object}	map[string]string			"Server error"
// @Router			/transactions/summary [get]
func (handler *Transaction_handler) Get_transactions_summary_handler(c *gin.Context) {
	//get the user claims.
	user, err := utils.Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//get transactions summary.
	summary, err := handler.service.Get_transactions_summary(user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(200, summary)
	case errors.Is(err, service.ErrWalletNotFound):
		c.JSON(404, gin.H{"error": "Wallet not found"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}

// @Summary	 		Creates a new category budget.
// @Description 	Sets a spending budget limit for a specific category for the current user.
// @Tags			Budget
// @Accept			json
// @Produce			json
// @Param			request body 	model.Create_budget_request true 	"Create budget details"
// @Success			201  {object}	model.Budget						"Created budget details"
// @Failure			400	 {object}	map[string]string					"Error message"
// @Failure			500	 {object}	map[string]string					"Server error"
// @Router			/budgets [post]
func (handler *Transaction_handler) Create_budget_handler(c *gin.Context) {
	//get the user claims.
	user, err := utils.Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//define a struct to get the request.
	var request model.Create_budget_request
	//get the request from body.
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid json"})
		return
	}
	//call service to perform operation.
	budget, err := handler.service.Create_budget(request, user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(201, budget)
	case errors.Is(err, service.ErrInvalidAmount):
		c.JSON(400, gin.H{"error": "Limit must be bigger than or equal zero"})
	case errors.Is(err, service.ErrEmptyCategory):
		c.JSON(400, gin.H{"error": "Category cannot be empty"})
	case errors.Is(err, service.ErrDuplicatedBudget):
		c.JSON(400, gin.H{"error": "Budget already exists for this user"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}

// @Summary	 		Updates an existing budget.
// @Description 	Updates a spending limit for a specific category for the current user.
// @Tags			Budget
// @Accept			json
// @Produce			json
// @Param			category path	string	true						"Budget category"
// @Param			request body 	model.Create_budget_request true 	"Update budget details"
// @Success			200  {object}	model.Budget						"Updated budget details"
// @Failure			400	 {object}	map[string]string					"Error message"
// @Failure			404	 {object}	map[string]string					"Budget not found"
// @Failure			500	 {object}	map[string]string					"Server error"
// @Router			/budgets/{category} [put]
func (handler *Transaction_handler) Update_budget_handler(c *gin.Context) {
	//get the user claims.
	user, err := utils.Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//define a struct to get the request.
	var request model.Create_budget_request
	//get the request from body.
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid json"})
		return
	}
	//call service to perform operation.
	budget, err := handler.service.Update_budget(request, user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(200, budget)
	case errors.Is(err, service.ErrInvalidAmount):
		c.JSON(400, gin.H{"error": "Limit must be bigger than or equal zero"})
	case errors.Is(err, service.ErrEmptyCategory):
		c.JSON(400, gin.H{"error": "Category cannot be empty"})
	case errors.Is(err, service.ErrBudgetNotFound):
		c.JSON(404, gin.H{"error": "Budget does not exists for this user"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}

// @Summary	 		Get budget status for all categories.
// @Description 	Get spending progress, monthly limits, and over-budget flags for all user budgets.
// @Tags			Budget
// @Produce			json
// @Success			200  {array}	model.Budget_status			"List of user budget statuses"
// @Failure			500	 {object}	map[string]string			"Server error"
// @Router			/budgets/status [get]
func (handler *Transaction_handler) Get_all_budget_status(c *gin.Context) {
	//get the user claims.
	user, err := utils.Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//get budget status for all user budgets.
	budgets, err := handler.service.Get_all_budget_status(user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(200, budgets)
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}