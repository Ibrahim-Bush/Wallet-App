package service

import (
	"Wallet-App/model"
	"Wallet-App/repository"
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

var (
	ErrInvalidUser      = errors.New("User claims cannot be empty")
	ErrWalletNotFound   = errors.New("Wallet not found")
	ErrUnathorizedUser  = errors.New("Unathorized user")
	ErrInvalidDate      = errors.New("Invalid date format")
	ErrInvalidDateRange = errors.New("Start date cannot be after end date")
	ErrEmptyCategory    = errors.New("Category cannot be empty")
	ErrInvalidAmount    = errors.New("Amount must be bigger than zero")
	ErrEmptyReceiver    = errors.New("Receiver cannot be empty")
	ErrReceiverNotFound = errors.New("Receiver not found")
	ErrInSufficient     = errors.New("The amount is bigger than balance")
	ErrInvalidTransfer  = errors.New("The sender is the same as the receiver")
)

type Wallet_service interface {
	Get_user_wallet(id int) (*model.Wallet, error)
	Get_all_wallets(user *model.User_claims) ([]model.Wallet, error)
	Deposit_process(request model.Transfer_request, user *model.User_claims) (*model.Transaction, error)
	Withdraw_process(request model.Transfer_request, user *model.User_claims) (*model.Transaction, error)
	Transfer_process(request model.Transfer_request, user *model.User_claims) (*model.Transaction, error)
	Get_user_transactions(user *model.User_claims) ([]model.Transaction, error)
	Get_transactions_by_category(category string, user *model.User_claims) ([]model.Transaction, error)
	Get_transactions_by_date(start, end string, user *model.User_claims) ([]model.Transaction, error)
	Get_transactions_summary(user *model.User_claims) ([]model.Transaction_summary, error)
}

type wallet_service struct {
	wallet_repo repository.Wallet_repo
	trans_repo  repository.Transaction_repo
	user_repo   repository.User_repo
	db          *gorm.DB
}

func Init_wallet_service(wallet repository.Wallet_repo, trans repository.Transaction_repo, user repository.User_repo, db *gorm.DB) Wallet_service {
	new_service := wallet_service{wallet_repo: wallet, trans_repo: trans, user_repo: user, db: db}
	return &new_service
}

func (service *wallet_service) Get_user_wallet(id int) (*model.Wallet, error) {
	//get the current user wallet.
	wallet, err := service.wallet_repo.Get_record_by_userID(id)
	//check the result.
	switch {
	case err == nil:
		return wallet, nil
	case errors.Is(err, repository.ErrWalletNotFound):
		return nil, ErrWalletNotFound
	default:
		return nil, ErrServerError
	}
}

func (service *wallet_service) Get_all_wallets(user *model.User_claims) ([]model.Wallet, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//check the user role.
	if user.Role != "admin" {
		return nil, ErrUnathorizedUser
	}
	//get the wallets.
	wallets, err := service.wallet_repo.Get_all_records()
	//check the result.
	switch {
	case err == nil:
		return wallets, nil
	default:
		return nil, ErrServerError
	}
}

func (service *wallet_service) Deposit_process(request model.Transfer_request, user *model.User_claims) (*model.Transaction, error) {
	//first check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//validate the request.
	err := Normalize_request_fields(&request)
	if err != nil {
		return nil, err
	}
	//define a variable for the transaction.
	var transaction_record *model.Transaction
	//start a transaction for deposit process.
	err = service.db.Transaction(func(tx *gorm.DB) error {
		//first get the wallet of the user and lock its row.
		wallet, err := service.wallet_repo.Get_record_by_userID_with_lock(tx, user.UserID)
		if err != nil {
			return err
		}
		//then apply the deposit request.
		wallet.Balance = wallet.Balance + request.Amount
		//store the change.
		err = service.wallet_repo.Update_wallet_balance(tx, wallet.ID, wallet.Balance)
		if err != nil {
			return err
		}
		//get the current time.
		current := time.Now().UTC()
		//record the transaction.
		transaction := model.Transaction{
			WalletID:        wallet.ID,
			Type:            "deposit",
			Amount:          request.Amount,
			Category:        request.Category,
			Note:            request.Note,
			RelatedWalletID: nil,
			CreatedAt:       current,
		}
		//store the transaction.
		err = service.trans_repo.Create_transaction_record(tx, &transaction)
		if err != nil {
			return err
		}
		//after successfull process.
		transaction_record = &transaction
		return nil
	})
	//check the result of process.
	switch {
	case err == nil:
		return transaction_record, nil
	case errors.Is(err, repository.ErrWalletNotFound):
		return nil, ErrWalletNotFound
	default:
		return nil, ErrServerError
	}
}

func (service *wallet_service) Withdraw_process(request model.Transfer_request, user *model.User_claims) (*model.Transaction, error) {
	//first check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//validate the request.
	err := Normalize_request_fields(&request)
	if err != nil {
		return nil, err
	}
	//define a variable for the transaction.
	var transaction_record *model.Transaction
	//start a transaction for withdraw process.
	err = service.db.Transaction(func(tx *gorm.DB) error {
		//first get the wallet of the user and lock its row.
		wallet, err := service.wallet_repo.Get_record_by_userID_with_lock(tx, user.UserID)
		if err != nil {
			return err
		}
		//then validate the withdraw request.
		if wallet.Balance < request.Amount {
			return ErrInSufficient
		}
		//apply the withdraw process.
		wallet.Balance = wallet.Balance - request.Amount
		//store the change.
		err = service.wallet_repo.Update_wallet_balance(tx, wallet.ID, wallet.Balance)
		if err != nil {
			return err
		}
		//get the current time.
		current := time.Now().UTC()
		//record the transaction.
		transaction := model.Transaction{
			WalletID:        wallet.ID,
			Type:            "withdraw",
			Amount:          request.Amount,
			Category:        request.Category,
			Note:            request.Note,
			RelatedWalletID: nil,
			CreatedAt:       current,
		}
		//store the transaction.
		err = service.trans_repo.Create_transaction_record(tx, &transaction)
		if err != nil {
			return err
		}
		//after successfull process.
		transaction_record = &transaction
		return nil
	})
	//check the result of process.
	switch {
	case err == nil:
		return transaction_record, nil
	case errors.Is(err, repository.ErrWalletNotFound):
		return nil, ErrWalletNotFound
	case errors.Is(err, ErrInSufficient):
		return nil, ErrInSufficient
	default:
		return nil, ErrServerError
	}
}

func (service *wallet_service) Transfer_process(request model.Transfer_request, user *model.User_claims) (*model.Transaction, error) {
	//first check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//validate the request.
	err := Normalize_request_fields(&request)
	if err != nil {
		return nil, err
	}
	//check the receiver.
	if request.ToUsername == "" {
		return nil, ErrEmptyReceiver
	}
	//define a variable for the transaction.
	var transaction_record *model.Transaction
	//start a transaction for deposit process.
	err = service.db.Transaction(func(tx *gorm.DB) error {
		//first get the wallet of the sernder and lock its row.
		sender_wallet, err := service.wallet_repo.Get_record_by_userID_with_lock(tx, user.UserID)
		if err != nil {
			return err
		}
		//validate the amount of transfer.
		if sender_wallet.Balance < request.Amount {
			return ErrInSufficient
		}
		//if valid, get the receiver wallet.
		receiver, err := service.user_repo.Get_user_record_by_name(request.ToUsername)
		if err != nil {
			return err
		}
		//get the receiver wallet and lock its row.
		receiver_wallet, err := service.wallet_repo.Get_record_by_userID_with_lock(tx, receiver.ID)
		if err != nil {
			if errors.Is(err, repository.ErrWalletNotFound) {
				return ErrReceiverNotFound
			}
			return err
		}
		//compare receiver wallet to sender wallet.
		if sender_wallet.ID == receiver_wallet.ID {
			return ErrInvalidTransfer
		}
		//then apply the transfer request: debit sender.
		sender_wallet.Balance = sender_wallet.Balance - request.Amount
		//store the change.
		err = service.wallet_repo.Update_wallet_balance(tx, sender_wallet.ID, sender_wallet.Balance)
		if err != nil {
			return err
		}
		//then credit the receiver.
		receiver_wallet.Balance = receiver_wallet.Balance + request.Amount
		//store the change.
		err = service.wallet_repo.Update_wallet_balance(tx, receiver_wallet.ID, receiver_wallet.Balance)
		if err != nil {
			if errors.Is(err, repository.ErrWalletNotFound) {
				return ErrReceiverNotFound
			}
			return err
		}
		//get the current time.
		current := time.Now().UTC()
		//record transfer out transaction.
		transfer_out_transaction := model.Transaction{
			WalletID:        sender_wallet.ID,
			Type:            "transfer_out",
			Amount:          request.Amount,
			Category:        request.Category,
			Note:            request.Note,
			RelatedWalletID: &receiver_wallet.ID,
			CreatedAt:       current,
		}
		//store the transaction.
		err = service.trans_repo.Create_transaction_record(tx, &transfer_out_transaction)
		if err != nil {
			return err
		}
		//record the transfer in transaction.
		transfer_in_transaction := model.Transaction{
			WalletID:        receiver_wallet.ID,
			Type:            "transfer_in",
			Amount:          request.Amount,
			Category:        request.Category,
			Note:            request.Note,
			RelatedWalletID: &sender_wallet.ID,
			CreatedAt:       current,
		}
		//store the transaction.
		err = service.trans_repo.Create_transaction_record(tx, &transfer_in_transaction)
		if err != nil {
			return err
		}
		//after successfull process.
		transaction_record = &transfer_out_transaction
		return nil
	})
	//check the result of process.
	switch {
	case err == nil:
		return transaction_record, nil
	case errors.Is(err, repository.ErrUserNotFound), errors.Is(err, ErrReceiverNotFound):
		return nil, ErrReceiverNotFound
	case errors.Is(err, ErrInSufficient):
		return nil, ErrInSufficient
	case errors.Is(err, ErrInvalidTransfer):
		return nil, ErrInvalidTransfer
	case errors.Is(err, repository.ErrWalletNotFound):
		return nil, ErrWalletNotFound
	default:
		return nil, ErrServerError
	}
}

func Normalize_request_fields(request *model.Transfer_request) error {
	//first check the amount.
	if request.Amount <= 0 {
		return ErrInvalidAmount
	}
	//noramlize the category.
	request.Category = strings.ToLower(strings.TrimSpace(request.Category))
	if request.Category == "" {
		return ErrEmptyCategory
	}
	//cut white sapces from note.
	request.Note = strings.TrimSpace(request.Note)
	//normalize receiver name.
	request.ToUsername = strings.ToLower(strings.TrimSpace(request.ToUsername))
	//if all fields normalized.
	return nil
}

func (service *wallet_service) Get_user_transactions(user *model.User_claims) ([]model.Transaction, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//get all user transactions.
	transactions, err := service.trans_repo.Get_all_records(user)
	//check the result.
	switch {
	case err == nil:
		return transactions, nil
	default:
		return nil, ErrServerError
	}
}

func (service *wallet_service) Get_transactions_by_category(category string, user *model.User_claims) ([]model.Transaction, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//normalize category.
	category = strings.ToLower(strings.TrimSpace(category))
	if category == "" {
		return nil, ErrEmptyCategory
	}
	//get all user transactions of the target category.
	transactions, err := service.trans_repo.Get_records_by_category(category, user)
	//check the result.
	switch {
	case err == nil:
		return transactions, nil
	default:
		return nil, ErrServerError
	}
}

func (service *wallet_service) Get_transactions_by_date(start, end string, user *model.User_claims) ([]model.Transaction, error) {
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
	//get all user transactions of the target category.
	transactions, err := service.trans_repo.Get_records_by_date(start_time, end_time, user)
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

func (service *wallet_service) Get_transactions_summary(user *model.User_claims) ([]model.Transaction_summary, error) {
	//check the user claims.
	if user == nil {
		return nil, ErrInvalidUser
	}
	//get the date of the current month.
	current_time := time.Now().UTC()
	current_month := time.Date(current_time.Year(), current_time.Month(), 1, 0, 0, 0, 0, time.UTC)
	//get the transactions summary of this month.
	summary, err := service.trans_repo.Get_records_summary(current_month, user)
	//check the summary.
	switch {
	case err == nil:
		return summary, nil
	default:
		return nil, ErrServerError
	}
}
