package mock

import (
	"Wallet-App/model"
	"time"

	"gorm.io/gorm"
)

type Trans_data struct {
	Received_transaction *model.Transaction
	Wanted_err           error
}

type Mock_transaction_repo struct {
	Create_transaction_calls []Trans_data
	Create_transaction_count int
}

func (mock *Mock_transaction_repo) Create_transaction_record(tx *gorm.DB, transaction *model.Transaction) error {
	//check if the call data was prepared.
	if len(mock.Create_transaction_calls) <= mock.Create_transaction_count {
		panic("Mock error: unexpected call to Create_transaction method")
	}
	//get the call data.
	call := &mock.Create_transaction_calls[mock.Create_transaction_count]
	//store the input id.
	call.Received_transaction = transaction
	//increase the calls count.
	mock.Create_transaction_count++
	//return the required valus.
	return call.Wanted_err
}

func (mock *Mock_transaction_repo) Get_all_records(wallet_id int, user *model.User_claims) ([]model.Transaction, error) {
	panic("Mock error: unexpected call to Get_all_transactions method")
}
func (mock *Mock_transaction_repo) Get_records_by_category(category string, wallet_id int, user *model.User_claims) ([]model.Transaction, error) {
	panic("Mock error: unexpected call to Get_transactions_by_category method")
}
func (mock *Mock_transaction_repo) Get_records_by_date(start, end time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction, error) {
	panic("Mock error: unexpected call to Get_transactions_by_date  method")
}
func (mock *Mock_transaction_repo) Get_records_summary(current_month time.Time, wallet_id int, user *model.User_claims) ([]model.Transaction_summary, error) {
	panic("Mock error: unexpected call to Get_transactions_summary method")
}
