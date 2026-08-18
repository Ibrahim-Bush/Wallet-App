package handler

import (
	"Wallet-App/model"
	"Wallet-App/service"
	"errors"

	"github.com/gin-gonic/gin"
)

var (
	ErrInvalidClaims = errors.New("Error invalid claims")
)

type Wallet_handler struct {
	service service.Wallet_service
}

func Init_wallet_handler(service service.Wallet_service) *Wallet_handler {
	new_handler := Wallet_handler{service: service}
	return &new_handler
}

func Get_user_claims(c *gin.Context) (*model.User_claims, error) {
	//get the user from gin context.
	value, exists := c.Get("user")
	if !exists {
		return nil, ErrInvalidClaims
	}
	//extract the user struct from interface.
	user, ok := value.(*model.User_claims)
	if !ok {
		return nil, ErrInvalidClaims
	}
	//if the struct extracted successfully.
	return user, nil
}

// @Summary	 		Get current user's wallet + balance.
// @Description 	Retrieve the current balance and details of the user wallet.
// @Tags			Wallet
// @Produce			json
// @Success			200  {object}	model.Wallet		"Wallet object"
// @Failure			403	 {object}	map[string]string	"Lack permissions"
// @Failure			404	 {object}	map[string]string	"Wallet not found"
// @Failure			500	 {object}	map[string]string	"Server error"
// @Router			/wallet [get]
func (handler *Wallet_handler) Get_wallet_handler(c *gin.Context) {
	//get the user claims.
	user, err := Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//check the role of the user.
	if user.Role == "admin" {
		wallets, err := handler.service.Get_all_wallets(user)
		switch {
		case err == nil:
			c.JSON(200, wallets)
		case errors.Is(err, service.ErrUnathorizedUser):
			c.JSON(403, gin.H{"error": "Lack required permissions"})
		default:
			c.JSON(500, gin.H{"error": "Server error"})
		}
		return
	}
	//if the user was not an admin.
	wallet, err := handler.service.Get_user_wallet(user.UserID)
	switch {
	case err == nil:
		c.JSON(200, *wallet)
	case errors.Is(err, service.ErrWalletNotFound):
		c.JSON(404, gin.H{"error": "wallet not found"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}

// @Summary	 		Creates a deposit transaction.
// @Description 	Deposit money into current user wallet and creates a transaction record.
// @Tags			Wallet
// @Accept			json
// @Produce			json
// @Param			request body 	model.Transfer_request true	"Deposit details"
// @Success			200  {object}	model.Transaction			"Transaction details"
// @Failure			400	 {object}	map[string]string			"Error message"
// @Failure			404	 {object}	map[string]string			"Wallet not found"
// @Failure			500	 {object}	map[string]string			"Server error"
// @Router			/wallet/deposit [post]
func (handler *Wallet_handler) Deposit_process_handler(c *gin.Context) {
	//get the user claims.
	user, err := Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//define a struct to get the request.
	var request model.Transfer_request
	//get the request from body.
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid json"})
		return
	}
	//call deposit service to perform operation.
	transaction, err := handler.service.Deposit_process(request, user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(200, transaction)
	case errors.Is(err, service.ErrInvalidAmount):
		c.JSON(400, gin.H{"error": "Transaction amount must be bigger than zero"})
	case errors.Is(err, service.ErrEmptyCategory):
		c.JSON(400, gin.H{"error": "Category cannot be empty"})
	case errors.Is(err, service.ErrWalletNotFound):
		c.JSON(404, gin.H{"error": "Wallet not found"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}

// @Summary	 		Creates a withdraw transaction.
// @Description 	withdraw money into current user wallet and creates a transaction record.
// @Tags			Wallet
// @Accept			json
// @Produce			json
// @Param			request body 	model.Transfer_request true "withdraw details"
// @Success			200  {object}	model.Transaction			"Transaction details"
// @Failure			400	 {object}	map[string]string			"Error message"
// @Failure			404	 {object}	map[string]string			"Wallet not found"
// @Failure			500	 {object}	map[string]string			"Server error"
// @Router			/wallet/withdraw [post]
func (handler *Wallet_handler) Withdraw_process_handler(c *gin.Context) {
	//get the user claims.
	user, err := Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//define a struct to get the request.
	var request model.Transfer_request
	//get the request from body.
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid json"})
		return
	}
	//call withdraw service to perform operation.
	transaction, err := handler.service.Withdraw_process(request, user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(200, transaction)
	case errors.Is(err, service.ErrInvalidAmount):
		c.JSON(400, gin.H{"error": "Transaction amount must be bigger than zero"})
	case errors.Is(err, service.ErrEmptyCategory):
		c.JSON(400, gin.H{"error": "Category cannot be empty"})
	case errors.Is(err, service.ErrInSufficient):
		c.JSON(400, gin.H{"error": "Insufficient balance for this transaction"})
	case errors.Is(err, service.ErrWalletNotFound):
		c.JSON(404, gin.H{"error": "Wallet not found"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
}

// @Summary	 		Creates a transfer transaction.
// @Description 	transfer money from current user's wallet to the receiver user's wallet.
// @Tags			Wallet
// @Accept			json
// @Produce			json
// @Param			request body 	model.Transfer_request true "Transfer details"
// @Success			200  {object}	model.Transaction			"Transaction details"
// @Failure			400	 {object}	map[string]string			"Error message"
// @Failure			404	 {object}	map[string]string			"Wallet or receiver not found"
// @Failure			500	 {object}	map[string]string			"Server error"
// @Router			/wallet/transfer [post]
func (handler *Wallet_handler) Transfer_process_handler(c *gin.Context) {
	//get the user claims.
	user, err := Get_user_claims(c)
	if err != nil {
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}
	//define a struct to get the request.
	var request model.Transfer_request
	//get the request from body.
	err = c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid json"})
		return
	}
	//call transfer service to perform operation.
	transaction, err := handler.service.Transfer_process(request, user)
	//check the result.
	switch {
	case err == nil:
		c.JSON(200, transaction)
	case errors.Is(err, service.ErrInvalidAmount):
		c.JSON(400, gin.H{"error": "Transaction amount must be bigger than zero"})
	case errors.Is(err, service.ErrEmptyCategory):
		c.JSON(400, gin.H{"error": "Category cannot be empty"})
	case errors.Is(err, service.ErrEmptyReceiver):
		c.JSON(400, gin.H{"error": "Receiver name cannot be empty"})
	case errors.Is(err, service.ErrInSufficient):
		c.JSON(400, gin.H{"error": "Insufficient balance for this transaction"})
	case errors.Is(err, service.ErrWalletNotFound):
		c.JSON(404, gin.H{"error": "Wallet not found"})
	case errors.Is(err, service.ErrReceiverNotFound):
		c.JSON(404, gin.H{"error": "Receiver not found"})
	case errors.Is(err, service.ErrInvalidTransfer):
		c.JSON(400, gin.H{"error": "The sender cannot be the receiver"})
	default:
		c.JSON(500, gin.H{"error": "Server error"})
	}
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
func (handler *Wallet_handler) Get_transactions_handler(c *gin.Context) {
	//get the user claims.
	user, err := Get_user_claims(c)
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
func (handler *Wallet_handler) Get_transactions_summary_handler(c *gin.Context) {
	//get the user claims.
	user, err := Get_user_claims(c)
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
