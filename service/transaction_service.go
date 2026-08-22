package service

import (
	"Wallet-App/model"
	"Wallet-App/repository"
	"errors"
	"strings"
	"time"
)

var (
	ErrBudgetNotFound   = errors.New("Error Budget record not found")
	ErrDuplicatedBudget = errors.New("Error budget already exist")
)

type Transaction_service interface {
	Get_user_transactions(user *model.User_claims) ([]model.Transaction, error)
	Get_transactions_by_category(category string, user *model.User_claims) ([]model.Transaction, error)
	Get_transactions_by_date(start, end string, user *model.User_claims) ([]model.Transaction, error)
	Get_transactions_summary(user *model.User_claims) ([]model.Transaction_summary, error)
	Create_budget(request model.Create_budget_request, user *model.User_claims) (*model.Budget, error)
	Update_budget(request model.Create_budget_request, user *model.User_claims) (*model.Budget, error)
	Get_all_budget_status(user *model.User_claims) ([]model.Budget_status, error)
}

type transaction_service struct {
	trans_repo  repository.Transaction_repo
	wallet_repo repository.Wallet_repo
}

func Init_transaction_service(trans repository.Transaction_repo, wallet_repo repository.Wallet_repo) Transaction_service {
	new_service := transaction_service{trans_repo: trans, wallet_repo: wallet_repo}
	return &new_service
}

func (service *transaction_service) Get_user_transactions(user *model.User_claims) ([]model.Transaction, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//first, get the user's wallet to get transactions on it.
	wallet, err := service.wallet_repo.Get_record_by_userID(user.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrWalletNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, ErrServerError
	}
	//get all user transactions.
	transactions, err := service.trans_repo.Get_all_records(wallet.ID, user)
	//check the result.
	switch {
	case err == nil:
		return transactions, nil
	default:
		return nil, ErrServerError
	}
}

func (service *transaction_service) Get_transactions_by_category(category string, user *model.User_claims) ([]model.Transaction, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//normalize category.
	category = strings.ToLower(strings.TrimSpace(category))
	if category == "" {
		return nil, ErrEmptyCategory
	}
	//first, get the user's wallet to get transactions on it.
	wallet, err := service.wallet_repo.Get_record_by_userID(user.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrWalletNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, ErrServerError
	}
	//get all user transactions of the target category.
	transactions, err := service.trans_repo.Get_records_by_category(category, wallet.ID, user)
	//check the result.
	switch {
	case err == nil:
		return transactions, nil
	default:
		return nil, ErrServerError
	}
}

func (service *transaction_service) Get_transactions_by_date(start, end string, user *model.User_claims) ([]model.Transaction, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//parse the standard dates.
	start_time, err := Parse_standard_date(start)
	//check the result.
	if err != nil {
		return nil, err
	}
	end_time, err := Parse_standard_date(end)
	if err != nil {
		return nil, err
	}
	//check if the end time is before start time.
	if start_time.After(end_time) {
		return nil, ErrInvalidDateRange
	}
	//first, get the user's wallet to get transactions on it.
	wallet, err := service.wallet_repo.Get_record_by_userID(user.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrWalletNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, ErrServerError
	}
	//get all user transactions of the target category.
	transactions, err := service.trans_repo.Get_records_by_date(start_time, end_time, wallet.ID, user)
	//check the result.
	switch {
	case err == nil:
		return transactions, nil
	default:
		return nil, ErrServerError
	}
}

func Parse_standard_date(input_date string) (time.Time, error) {
	//first, clean empty spaces at the start and end.
	clean_date := strings.TrimSpace(input_date)
	if clean_date == "" {
		return time.Time{}, ErrInvalidDate
	}
	//convert string to standard date.
	standard_date, err := time.Parse(time.RFC3339, clean_date)
	if err != nil {
		//try to convert it to date only without time.
		standard_date, err = time.Parse("2006-01-02", clean_date)
		if err != nil {
			return time.Time{}, ErrInvalidDate
		}
	}
	//convert date to UTC.
	standard_date = standard_date.UTC()
	//return standard time.
	return standard_date, nil
}

func (service *transaction_service) Get_transactions_summary(user *model.User_claims) ([]model.Transaction_summary, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//get the date of the current month.
	current_time := time.Now().UTC()
	current_month := time.Date(current_time.Year(), current_time.Month(), 1, 0, 0, 0, 0, time.UTC)
	//first, get the user's wallet to get transactions on it.
	wallet, err := service.wallet_repo.Get_record_by_userID(user.UserID)
	if err != nil {
		if errors.Is(err, repository.ErrWalletNotFound) {
			return nil, ErrWalletNotFound
		}
		return nil, ErrServerError
	}
	//get the transactions summary of this month.
	summary, err := service.trans_repo.Get_records_summary(current_month, wallet.ID, user)
	//check the summary.
	switch {
	case err == nil:
		return summary, nil
	default:
		return nil, ErrServerError
	}
}

func (service *transaction_service) Create_budget(request model.Create_budget_request, user *model.User_claims) (*model.Budget, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//check the input model.
	err := validate_budget_request(&request)
	if err != nil {
		return nil, err
	}
	//create a budget struct to store it.
	var new_budget model.Budget
	//store request fields in the struct.
	new_budget.Category = request.Category
	new_budget.MonthlyLimit = request.MonthlyLimit
	new_budget.UserID = user.UserID
	//call repo to store the budget record.
	err = service.trans_repo.Create_budget_record(&new_budget)
	//check the err.
	switch {
	case err == nil:
		return &new_budget, nil
	case errors.Is(err, repository.ErrDuplicatedBudget):
		return nil, ErrDuplicatedBudget
	default:
		return nil, ErrServerError
	}
}

func validate_budget_request(request *model.Create_budget_request) error {
	//Validate category field.
	request.Category = strings.ToLower(strings.TrimSpace(request.Category))
	if request.Category == "" {
		return ErrEmptyCategory
	}
	//validate the limit field.
	if request.MonthlyLimit < 0 {
		return ErrInvalidAmount
	}
	return nil
}

func (service *transaction_service) Update_budget(request model.Create_budget_request, user *model.User_claims) (*model.Budget, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//check the input model.
	err := validate_budget_request(&request)
	if err != nil {
		return nil, err
	}
	//get the stored record to update it.
	budget, err := service.trans_repo.Get_budget_record(user.UserID, request.Category)
	//check the error.
	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return nil, ErrBudgetNotFound
		}
		return nil, ErrServerError
	}
	//update the limit.
	budget.MonthlyLimit = request.MonthlyLimit
	//store the updated record.
	err = service.trans_repo.Update_budget_record(budget)
	switch {
	case err == nil:
		return budget, nil
	default:
		return nil, ErrServerError
	}
}

func (service *transaction_service) Get_all_budget_status(user *model.User_claims) ([]model.Budget_status, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//first get all budget records.
	budgets, err := service.trans_repo.Get_all_budget_records(user.UserID)
	//check err.
	if err != nil {
		return nil, ErrServerError
	}
	//then get all transactions summary for user.
	summary, err := service.Get_transactions_summary(user)
	//check the err.
	if err != nil {
		return nil, ErrServerError
	}
	//now, convert the summary to map to avoid nested loop.
	summary_map := make(map[string]int, len(summary))
	for _, item := range summary {
		summary_map[item.Category] = item.Total
	}
	//finally create the budget status for all budgets.
	budget_status := make([]model.Budget_status, 0, len(budgets))
	for _, budget := range budgets {
		spent := summary_map[budget.Category]
		//create a budget status.
		status := model.Budget_status{
			Category:     budget.Category,
			MonthlyLimit: budget.MonthlyLimit,
			Spent:        spent,
			OverBudget:   spent > budget.MonthlyLimit,
		}
		budget_status = append(budget_status, status)
	}
	//return the budget status slice.
	return budget_status, nil
}
