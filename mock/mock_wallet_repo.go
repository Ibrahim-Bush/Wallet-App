package mock

import (
	"Wallet-App/model"

	"gorm.io/gorm"
)

type Data struct {
	Received_id      int
	Received_balance int
	Wanted_wallet    *model.Wallet
	Wanted_err       error
}
type Mock_wallet_repo struct {
	Get_by_id_calls []Data
	Get_by_id_count int
	Update_calls  []Data
	Update_count  int
	Received_wallet *model.Wallet
	Wanted_err error
	Was_called bool
}

func (mock *Mock_wallet_repo) Get_record_by_userID_with_lock(tx *gorm.DB, id int) (*model.Wallet, error) {
	//check if the call data was prepared.
	if len(mock.Get_by_id_calls) <= mock.Get_by_id_count {
		panic("Mock error: unexpected call to Get_wallet_by_id_with_lock method")
	}
	//get the call data.
	call := &mock.Get_by_id_calls[mock.Get_by_id_count]
	//store the input id.
	call.Received_id = id
	//increase the calls count.
	mock.Get_by_id_count++
	//return the required valus.
	return call.Wanted_wallet, call.Wanted_err
}

func (mock *Mock_wallet_repo) Update_wallet_balance(tx *gorm.DB, wallet_id, new_balance int) error {
	//check if the call data was prepared.
	if len(mock.Update_calls) <= mock.Update_count {
		panic("Mock error: unexpected call to Update_wallet_balance method")
	}
	//get the call data.
	call := &mock.Update_calls[mock.Update_count]
	//store the input data
	call.Received_id = wallet_id
	call.Received_balance = new_balance
	//increase the calls count.
	mock.Update_count++
	//return the required valus.
	return call.Wanted_err
}

func (mock *Mock_wallet_repo) Create_wallet_record(tx *gorm.DB, new_wallet *model.Wallet) error {
	//change was called to true.
	mock.Was_called = true
	//store the received wallet.
	mock.Received_wallet = new_wallet
	//return the required error.
	return mock.Wanted_err

}
func (mock *Mock_wallet_repo) Get_record_by_userID(id int) (*model.Wallet, error) {
	panic("Mock error: unexpected call to Get_wallet_by_id method")
}
func (mock *Mock_wallet_repo) Get_all_records() ([]model.Wallet, error) {
	panic("Mock error: unexpected call to Get_all_wallets method")
}
