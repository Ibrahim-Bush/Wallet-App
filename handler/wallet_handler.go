package handler

import (
	"Wallet-App/model"
	"Wallet-App/service"
	"Wallet-App/utils"
	"errors"

	"github.com/gin-gonic/gin"
)


type Wallet_handler struct {
	service service.Wallet_service
}

func Init_wallet_handler(service service.Wallet_service) *Wallet_handler {
	new_handler := Wallet_handler{service: service}
	return &new_handler
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
	user, err := utils.Get_user_claims(c)
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
	user, err := utils.Get_user_claims(c)
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
// @Description 	withdraw money from current user wallet, creates a transaction record and display a budget warning.
// @Tags			Wallet
// @Accept			json
// @Produce			json
// @Param			request body 	model.Transfer_request true "withdraw details"
// @Success			200  {object}	model.Transaction_response	"Transaction details with budget warning"
// @Failure			400	 {object}	map[string]string			"Error message"
// @Failure			404	 {object}	map[string]string			"Wallet not found"
// @Failure			500	 {object}	map[string]string			"Server error"
// @Router			/wallet/withdraw [post]
func (handler *Wallet_handler) Withdraw_process_handler(c *gin.Context) {
	//get the user claims.
	user, err := utils.Get_user_claims(c)
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
// @Success			200  {object}	model.Transaction_response	"Transaction details with budget warning"
// @Failure			400	 {object}	map[string]string			"Error message"
// @Failure			404	 {object}	map[string]string			"Wallet or receiver not found"
// @Failure			500	 {object}	map[string]string			"Server error"
// @Router			/wallet/transfer [post]
func (handler *Wallet_handler) Transfer_process_handler(c *gin.Context) {
	//get the user claims.
	user, err := utils.Get_user_claims(c)
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
