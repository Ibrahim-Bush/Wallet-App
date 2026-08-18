package main

import (
	database "Wallet-App/config"
	_ "Wallet-App/docs"
	"Wallet-App/handler"
	"Wallet-App/repository"
	"Wallet-App/router"
	"Wallet-App/service"
	"log"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title			Wallet & Expense Tracker API
// @version 		1.0
// @description 	Backend system for user wallet, balance and transaction handling.
// @host			localhost:8080
// @BasePath		/

func main() {

	//init new router.
	Router := gin.Default()

	//add the swagger documentation on router.
	Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	//then: get a connection to database.
	connection_data := "host=db user=postgres password=admin dbname=wallet_app port=5432 sslmode=disable"
	database, err := database.Init_database(connection_data)
	//if something went wrong exit.
	if err != nil {
		log.Fatal(err)
	}

	//linking database to repository layer.
	user_repo := repository.Init_user_repo(database)
	wallet_repo := repository.Init_wallet_repo(database)
	transaction_repo := repository.Init_transaction_repo(database)

	//linking repository layer with service layer.
	user_service := service.Init_user_service(user_repo, wallet_repo, database)
	wallet_service := service.Init_wallet_service(wallet_repo, transaction_repo, user_repo, database)

	//linking service layer to handler layer.
	user_handler := handler.Init_user_handler(user_service)
	wallet_handler := handler.Init_wallet_handler(wallet_service)

	//linking the handler layer to the router.
	router.Init_user_router(Router, user_handler)
	router.Init_wallet_router(Router, wallet_handler)

	//run the server at local port ":8080".
	Router.Run(":8080")

}
