package router

import (
	"Wallet-App/handler"
	"Wallet-App/middleware"

	"github.com/gin-gonic/gin"
)

func Init_wallet_router(router *gin.Engine, handler *handler.Wallet_handler) {

	//create a routes group for wallets.
	walletAPI := router.Group("/wallet")

	//apply authentication middleware to wallet group.
	walletAPI.Use(middleware.Auth_middleware)

	walletAPI.GET("", handler.Get_wallet_handler)
	walletAPI.POST("/deposit", middleware.Authorize_user_middleware, handler.Deposit_process_handler)
	walletAPI.POST("/withdraw", middleware.Authorize_user_middleware, handler.Withdraw_process_handler)
	walletAPI.POST("/transfer", middleware.Authorize_user_middleware, handler.Transfer_process_handler)

}
