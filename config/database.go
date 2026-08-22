package database

import (
	"Wallet-App/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Init_database(connection_data string) (*gorm.DB, error) {

	//first get connection to the database.
	db, err := gorm.Open(postgres.Open(connection_data), &gorm.Config{TranslateError: true})
	//check if something went wrong.
	if err != nil {
		return nil, err
	}
	//create a table with AutoMigrate.
	err = db.AutoMigrate(&model.User{}, &model.Wallet{}, &model.Transaction{}, &model.Budget{})
	if err != nil {
		return nil, err
	}
	//if the connection and autoMigration succeeded, return the database object.
	return db, nil
}
