package repository

import (
	"Wallet-App/model"
	"errors"
	"time"

	"gorm.io/gorm"
)

var (
	ErrRecordNotFound   = errors.New("Budget record not found")
	ErrDuplicatedBudget = errors.New("Budget already exists")
)

type Transaction_repo interface {
	Get_all_records(wallet_id int, user *model.User_claims) ([]model.Transaction, error)
	Get_records_by_category(category string, wallet_id int, user *model.User_claims) ([]model.Transaction, error)
	Get_records_by_date(start, end time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction, error)
	Get_records_summary(current_month time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction_summary, error)
	Get_category_summary(category string, current_month time.Time, wallet_id int) (*model.Transaction_summary, error)
	Create_transaction_record(tx *gorm.DB, transaction *model.Transaction) error
	Create_budget_record(budget *model.Budget) error
	Update_budget_record(budget *model.Budget) error
	Get_budget_record(user_id int, category string) (*model.Budget, error)
	Get_all_budget_records(user_id int) ([]model.Budget, error)
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

func (t *transaction_repo) Get_records_by_date(start, end time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction, error) {
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

func (t *transaction_repo) Get_category_summary(category string, current_month time.Time, wallet_id int) (*model.Transaction_summary, error) {
	//define a variable for the category total spent.
	var total int
	//get total spent in this category.
	result := t.db.Model(&model.Transaction{}).Select("SUM(amount)").
		Where("created_at >= ? AND category = ? AND wallet_id = ?", current_month, category, wallet_id).Scan(&total)
	//set the struct with summary.
	summary := model.Transaction_summary{
		Category: category,
		Total:    total,
	}
	//check the result.
	switch {
	case result.Error == nil:
		return &summary, nil
	default:
		return nil, result.Error
	}
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

func (t *transaction_repo) Create_budget_record(budget *model.Budget) error {
	//create a new record.
	result := t.db.Create(budget)
	//check result of process.
	switch {
	case errors.Is(result.Error, gorm.ErrDuplicatedKey):
		return ErrDuplicatedBudget
	default:
		return result.Error
	}
}

func (t *transaction_repo) Update_budget_record(budget *model.Budget) error {
	//update the record.
	result := t.db.Save(budget)
	//reutrn the result of process.
	return result.Error
}

func (t *transaction_repo) Get_budget_record(user_id int, category string) (*model.Budget, error) {
	//define a model for the record.
	var record model.Budget
	//get the required record.
	result := t.db.Where("category = ? AND user_id = ?", category, user_id).First(&record)
	//check the result of process.
	switch {
	case result.Error == nil:
		return &record, nil
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return nil, ErrRecordNotFound
	default:
		return nil, result.Error
	}
}

func (t *transaction_repo) Get_all_budget_records(user_id int) ([]model.Budget, error) {
	//define a slice for records.
	var records []model.Budget
	//get all budget records.
	result := t.db.Where("user_id = ?", user_id).Find(&records)
	//check the result of process.
	if result.Error == nil {
		return records, nil
	}
	return nil, result.Error
}
