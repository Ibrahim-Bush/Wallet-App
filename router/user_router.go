package router

import (
	"Wallet-App/handler"

	"github.com/gin-gonic/gin"
)

func Init_user_router(router *gin.Engine, handler *handler.User_handler) {

	//register user routes.
	router.POST("/signup", handler.Create_user_handler)
	router.POST("/login", handler.Login_user_handler)

}
