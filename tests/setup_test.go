package test

import (
	database "Wallet-App/config"
	"Wallet-App/handler"
	"Wallet-App/model"
	"Wallet-App/repository"
	"Wallet-App/router"
	"Wallet-App/service"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var (
	Router     *gin.Engine
	DB         *gorm.DB
	user1_Auth = model.Auth_request{Username: "		ALi		", Password: "	admin1	"}
	user2_Auth = model.Auth_request{Username: "		MoHameD		", Password: "	admin2	"}
)

func TestMain(m *testing.M) {

	//init the testing database.
	connection_data := "host=localhost user=postgres password=admin dbname=wallet_app_test port=5432 sslmode=disable"
	//connect to the database.
	database, err := database.Init_database(connection_data)
	if err != nil {
		panic("Failed to setup testing database")
	}
	//if succeeded assign it to global variable.
	DB = database

	//init the testing router.
	Router = gin.Default()

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

	//setup the testing users.
	Create_test_user(user1_Auth)
	Create_test_user(user2_Auth)

	//run the test.
	result := m.Run()

	//clear the database after testing.
	database.Migrator().DropTable(&model.Transaction{}, &model.Wallet{}, &model.User{})
	//get the result of testing in terminal.
	os.Exit(result)

}
