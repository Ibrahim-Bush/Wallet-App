package router

import (
	"Wallet-App/handler"
	"Wallet-App/middleware"

	"github.com/gin-gonic/gin"
)

func Init_transaction_router(router *gin.Engine, handler *handler.Transaction_handler) {

	//create a routes group for transactions.
	transactionAPI := router.Group("/transactions")

	//apply authentication middleware to transactions group.
	transactionAPI.Use(middleware.Auth_middleware)

	transactionAPI.GET("", handler.Get_transactions_handler)
	transactionAPI.GET("/summary", handler.Get_transactions_summary_handler)

	//create a routes group for budgets.
	BudgetAPI := router.Group("/budgets")

	//apply authentication middleware to transactions group.
	BudgetAPI.Use(middleware.Auth_middleware, middleware.Authorize_user_middleware)

	BudgetAPI.POST("", handler.Create_budget_handler)
	BudgetAPI.PUT("/:category", handler.Update_budget_handler)
	BudgetAPI.GET("/status", handler.Get_all_budget_status)

}
