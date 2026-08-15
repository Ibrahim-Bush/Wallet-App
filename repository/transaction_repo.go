package repository

import (
	"Wallet-App/model"
	"time"

	"gorm.io/gorm"
)

type Transaction_repo interface {
	Get_all_records(wallet_id int, user *model.User_claims) ([]model.Transaction, error)
	Get_records_by_category(category string, wallet_id int, user *model.User_claims) ([]model.Transaction, error)
	Get_records_by_date(start, end time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction, error)
	Get_records_summary(current_month time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction_summary, error)
	Create_transaction_record(tx *gorm.DB, transaction *model.Transaction) error
}

func Init_transaction_repo(db *gorm.DB) Transaction_repo {
	new_repo := transaction_repo{db: db}
	return &new_repo
}

type transaction_repo struct {
	db *gorm.DB
}

func (t *transaction_repo) Get_all_records(wallet_id int, user *model.User_claims) ([]model.Transaction, error) {
	//create a slice for elements.
	var transactions = make([]model.Transaction, 0)
	//define a variable to check the result.
	var result *gorm.DB
	//differentiate users based on role.
	if user.Role == "admin" {
		//get all records.
		result = t.db.Find(&transactions)
	} else {
		//get records of the current user only.
		result = t.db.Where("wallet_id = ?", wallet_id).Find(&transactions)
	}
	//check the result of process.
	if result.Error != nil {
		//return empty slice and error value.
		return nil, result.Error
	}
	return transactions, nil
}

func (t *transaction_repo) Get_records_by_category(category string, wallet_id int, user *model.User_claims) ([]model.Transaction, error) {
	//create a slice for elements.
	var transactions = make([]model.Transaction, 0)
	//define a variable to check the result.
	var result *gorm.DB
	//differentiate users based on role.
	if user.Role == "admin" {
		//get all records of the target category.
		result = t.db.Where("category = ?", category).Find(&transactions)
	} else {
		//get records of the current user only.
		result = t.db.Where("category = ? AND wallet_id = ?", category, wallet_id).Find(&transactions)
	}
	//check the result of process.
	if result.Error != nil {
		//return empty slice and error value.
		return nil, result.Error
	}
	return transactions, nil
}

func (t *transaction_repo) Get_records_by_date( start, end time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction, error) {
	//create a slice for elements.
	var transactions = make([]model.Transaction, 0)
	//define a variable to check the result.
	var result *gorm.DB
	//differentiate users based on role.
	if user.Role == "admin" {
		//get all records between start and end dates.
		result = t.db.Where("created_at BETWEEN ? AND ?", start, end).Find(&transactions)
	} else {
		//get records of the current user only.
		result = t.db.Where("wallet_id = ? AND created_at BETWEEN ? AND ?", wallet_id, start, end).Find(&transactions)
	}
	//check the result of process.
	if result.Error != nil {
		//return empty slice and error value.
		return nil, result.Error
	}
	return transactions, nil
}

func (t *transaction_repo) Get_records_summary(current_month time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction_summary, error) {
	//create a slice for elements.
	var summary = make([]model.Transaction_summary, 0)
	//define a variable to check the result.
	var result *gorm.DB
	//differentiate users based on role.
	if user.Role == "admin" {
		result = t.db.Model(&model.Transaction{}).Select("category, SUM(amount) as total").
			Where("created_at >= ?", current_month).Group("category").Scan(&summary)
	} else {
		result = t.db.Model(&model.Transaction{}).Select("category, SUM(amount) as total").
			Where("wallet_id = ? AND created_at >= ?", wallet_id, current_month).Group("category").Scan(&summary)
	}
	//check the result of process.
	if result.Error != nil {
		//return empty slice and error value.
		return nil, result.Error
	}
	return summary, nil
}

func (t *transaction_repo) Create_transaction_record(tx *gorm.DB, transaction *model.Transaction) error {
	//check the transaction object.
	if tx == nil {
		tx = t.db
	}
	//create the transaction.
	result := tx.Create(transaction)
	//check the result.
	return result.Error
}
