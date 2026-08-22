package service

import (
	"Wallet-App/mock"
	"Wallet-App/model"
	"Wallet-App/repository"
	"Wallet-App/utils"
	"errors"
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func Setup_testing_db() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to in-memory test database")
	}
	return db
}

func Test_get_user_wallet(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name string

		input_id          int
		repo_id           int
		get_wallet_return []mock.Data
		expected_wallet   *model.Wallet
		expected_err      error
	}
	//set a table of tests.
	tests := []usecase{
		{
			name:              "Error - Non-existent wallet",
			input_id:          1,
			repo_id:           1,
			get_wallet_return: []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			expected_wallet:   nil,
			expected_err:      ErrWalletNotFound,
		}, {
			name:              "Error - Server error with get wallet",
			input_id:          1,
			repo_id:           1,
			get_wallet_return: []mock.Data{{Wanted_wallet: nil, Wanted_err: errors.New("Database connection failure")}},
			expected_wallet:   nil,
			expected_err:      ErrServerError,
		}, {
			name:              "Success - Get wallet successfully",
			input_id:          1,
			repo_id:           1,
			get_wallet_return: []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2, UserID: 1}, Wanted_err: nil}},
			expected_wallet:   &model.Wallet{ID: 2, UserID: 1},
			expected_err:      nil,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the get_wallet service.
			wallet, err := service.Get_user_wallet(usecase.input_id)
			//check the response.
			if !errors.Is(err, usecase.expected_err) {
				t.Errorf("Response: expected error %v, got %v", usecase.expected_err, err)
			}
			//check the wallet in response.
			if usecase.expected_wallet == nil && wallet != nil {
				t.Errorf("Response: expected nil wallet got %+v wallet", wallet)
			} else if usecase.expected_wallet != nil && wallet == nil {
				t.Errorf("Response: expected not nil wallet got %v wallet", wallet)
			}
			//check the passed user id.
			received_id := mock_wallet_repo.Get_by_id_calls[0].Received_id
			if received_id != usecase.repo_id {
				t.Errorf("Logic: expected user ID passed to get_wallet_by_userID %d got %d", usecase.repo_id, received_id)
			}
		})
	}
}

func Test_get_all_wallets(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name          string
		input_user    *model.User_claims
		repo_list     []model.Wallet
		repo_err      error
		expected_list []model.Wallet
		expected_err  error
		repo_call     bool
	}
	//set a table of usecase.
	tests := []usecase{
		{
			name:          "Error: Invalid user claims",
			input_user:    nil,
			expected_list: nil,
			expected_err:  ErrInvalidUser,
			repo_call:     false,
		}, {
			name:          "Error: Unauthorized user",
			input_user:    &model.User_claims{Role: "user"},
			expected_list: nil,
			expected_err:  ErrUnathorizedUser,
			repo_call:     false,
		}, {
			name:          "Error - Server error with get wallets",
			input_user:    &model.User_claims{Role: "admin"},
			repo_list:     nil,
			repo_err:      errors.New("Database connection failure"),
			expected_list: nil,
			expected_err:  ErrServerError,
			repo_call:     true,
		}, {
			name:          "Success - Get wallets successfully",
			input_user:    &model.User_claims{Role: "admin"},
			repo_list:     []model.Wallet{{ID: 1}, {ID: 2}, {ID: 3}},
			repo_err:      nil,
			expected_list: []model.Wallet{{ID: 1}, {ID: 2}, {ID: 3}},
			expected_err:  nil,
			repo_call:     true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Wanted_list: usecase.repo_list, Wanted_err: usecase.repo_err}
			mock_trans_repo := &mock.Mock_transaction_repo{}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the get_all_wallets service.
			wallets, err := service.Get_all_wallets(usecase.input_user)
			//check the response.
			if !errors.Is(err, usecase.expected_err) {
				t.Errorf("Response: expected error %v, got %v", usecase.expected_err, err)
			}
			//check the list in response.
			if usecase.expected_list == nil && wallets != nil {
				t.Errorf("Response: expected nil slice of wallets got %+v wallets", wallets)
			} else if expected := len(usecase.expected_list); expected != len(wallets) {
				t.Errorf("Response: expected slice of length %d got slice of length %d", expected, len(wallets))
			}
			//check the repo call.
			if mock_wallet_repo.Was_called != usecase.repo_call {
				t.Errorf("Behaviour: expected repo call %t got %t", usecase.repo_call, mock_wallet_repo.Was_called)
			}
		})
	}
}

func Test_deposit_process(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name                 string
		input_request        model.Transfer_request
		input_user           *model.User_claims
		get_by_id_return     []mock.Data
		update_wallet_return []mock.Data
		repo_balance         int
		create_trans_return  []mock.Trans_data
		repo_transaction     *model.Transaction
		expected_transaction *model.Transaction
		expected_err         error
		get_by_id_calls      int
		update_calls         int
		create_trans_calls   int
	}
	//create a table of usecase.
	tests := []usecase{
		{
			name:                 "Error: Invalid transaction amount",
			input_request:        model.Transfer_request{Amount: -1, Category: "		"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrInvalidAmount,
			get_by_id_calls:      0,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:                 "Error: Empty category",
			input_request:        model.Transfer_request{Amount: 1, Category: "		"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrEmptyCategory,
			get_by_id_calls:      0,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:                 "Error: Wallet not found",
			input_request:        model.Transfer_request{Amount: 1, Category: "	FOOd	"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrWalletNotFound,
			get_by_id_calls:      1,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:          "Error: Server error with update balance",
			input_request: model.Transfer_request{Amount: 1, Category: "	FOOd	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 0, ID: 1, UserID: 1},
				Wanted_err: nil}},
			update_wallet_return: []mock.Data{{Wanted_err: errors.New("Database connection failure")}},
			repo_balance:         1,
			create_trans_return:  []mock.Trans_data{},
			repo_transaction:     nil,
			expected_transaction: nil,
			expected_err:         ErrServerError,
			get_by_id_calls:      1,
			update_calls:         1,
			create_trans_calls:   0,
		}, {
			name:          "Error: Server error with create transaction record",
			input_request: model.Transfer_request{Amount: 10, Category: "	FOOd	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 0, ID: 1, UserID: 1},
				Wanted_err: nil}},
			update_wallet_return: []mock.Data{{Wanted_err: nil}},
			repo_balance:         10,
			create_trans_return:  []mock.Trans_data{{Wanted_err: errors.New("Database connection failure")}},
			repo_transaction:     &model.Transaction{WalletID: 1, Type: "deposit", Amount: 10, Category: "food", RelatedWalletID: nil},
			expected_transaction: nil,
			expected_err:         ErrServerError,
			get_by_id_calls:      1,
			update_calls:         1,
			create_trans_calls:   1,
		}, {
			name:          "Success: Transaction created",
			input_request: model.Transfer_request{Amount: 10, Category: "	FOOd	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 30, ID: 1, UserID: 1},
				Wanted_err: nil}},
			update_wallet_return: []mock.Data{{Wanted_err: nil}},
			repo_balance:         40,
			create_trans_return:  []mock.Trans_data{{Wanted_err: nil}},
			repo_transaction:     &model.Transaction{WalletID: 1, Type: "deposit", Amount: 10, Category: "food", RelatedWalletID: nil},
			expected_transaction: &model.Transaction{WalletID: 1, Type: "deposit", Amount: 10, Category: "food", RelatedWalletID: nil},
			expected_err:         nil,
			get_by_id_calls:      1,
			update_calls:         1,
			create_trans_calls:   1,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_by_id_return, Update_calls: usecase.update_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Create_transaction_calls: usecase.create_trans_return}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the deposit service.
			transaction, err := service.Deposit_process(usecase.input_request, usecase.input_user)
			//check the response.
			if !errors.Is(err, usecase.expected_err) {
				t.Errorf("Response: expected error %v, got %v", usecase.expected_err, err)
			}
			//check the repo calls.
			if mock_wallet_repo.Get_by_id_count != usecase.get_by_id_calls {
				t.Errorf("Behaviour: expected get_wallet_by_id calls %d, got %d", usecase.get_by_id_calls, mock_wallet_repo.Get_by_id_count)
			}
			if mock_wallet_repo.Update_count != usecase.update_calls {
				t.Errorf("Behaviour: expected update_wallet_balance calls %d, got %d", usecase.update_calls, mock_wallet_repo.Update_count)
			}
			if mock_trans_repo.Create_transaction_count != usecase.create_trans_calls {
				t.Errorf("Behaviour: expected create_transaction calls %d, got %d", usecase.create_trans_calls, mock_trans_repo.Create_transaction_count)
			}
			//check the repo input.
			if usecase.update_calls > 0 && len(mock_wallet_repo.Update_calls) > 0 {
				//check the new balance.
				received_balance := mock_wallet_repo.Update_calls[0].Received_balance
				if received_balance != usecase.repo_balance {
					t.Errorf("Logic: expected new wallet balance in repo %d got %d", usecase.repo_balance, received_balance)
				}
			}
			if usecase.create_trans_calls > 0 && len(mock_trans_repo.Create_transaction_calls) > 0 {
				//check the new balance.
				received_transaction := mock_trans_repo.Create_transaction_calls[0].Received_transaction
				//compare received with expected.
				utils.Compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_transaction)
			}
			//comare the output transaction to expected.
			utils.Compare_transaction_structs(t, "Response", transaction, usecase.expected_transaction)
		})
	}
}

func Test_withdraw_process(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name                 string
		input_request        model.Transfer_request
		input_user           *model.User_claims
		get_by_id_return     []mock.Data
		update_wallet_return []mock.Data
		repo_balance         int
		create_trans_return  []mock.Trans_data
		repo_transaction     *model.Transaction
		expected_transaction *model.Transaction
		expected_err         error
		get_by_id_calls      int
		update_calls         int
		create_trans_calls   int
	}
	//create a table of usecase.
	tests := []usecase{
		{
			name:                 "Error: Invalid transaction amount",
			input_request:        model.Transfer_request{Amount: -10, Category: "		"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrInvalidAmount,
			get_by_id_calls:      0,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:                 "Error: Empty category",
			input_request:        model.Transfer_request{Amount: 1, Category: "		"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrEmptyCategory,
			get_by_id_calls:      0,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:                 "Error: Wallet not found",
			input_request:        model.Transfer_request{Amount: 1, Category: "	FOOd	"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrWalletNotFound,
			get_by_id_calls:      1,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:          "Error: Insufficient balance",
			input_request: model.Transfer_request{Amount: 20, Category: "	FOOd	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 10, ID: 1, UserID: 1},
				Wanted_err: nil}},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrInSufficient,
			get_by_id_calls:      1,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:          "Error: Server error with update balance",
			input_request: model.Transfer_request{Amount: 30, Category: "	FOOd	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1},
				Wanted_err: nil}},
			update_wallet_return: []mock.Data{{Wanted_err: errors.New("Database connection failure")}},
			repo_balance:         70,
			create_trans_return:  []mock.Trans_data{},
			repo_transaction:     nil,
			expected_transaction: nil,
			expected_err:         ErrServerError,
			get_by_id_calls:      1,
			update_calls:         1,
			create_trans_calls:   0,
		}, {
			name:          "Error: Server error with create transaction record",
			input_request: model.Transfer_request{Amount: 10, Category: "	FOOd	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 50, ID: 1, UserID: 1},
				Wanted_err: nil}},
			update_wallet_return: []mock.Data{{Wanted_err: nil}},
			repo_balance:         40,
			create_trans_return:  []mock.Trans_data{{Wanted_err: errors.New("Database connection failure")}},
			repo_transaction:     &model.Transaction{WalletID: 1, Type: "withdraw", Amount: 10, Category: "food", RelatedWalletID: nil},
			expected_transaction: nil,
			expected_err:         ErrServerError,
			get_by_id_calls:      1,
			update_calls:         1,
			create_trans_calls:   1,
		}, {
			name:          "Success: Transaction created",
			input_request: model.Transfer_request{Amount: 10, Category: "	FOOd	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 30, ID: 1, UserID: 1},
				Wanted_err: nil}},
			update_wallet_return: []mock.Data{{Wanted_err: nil}},
			repo_balance:         20,
			create_trans_return:  []mock.Trans_data{{Wanted_err: nil}},
			repo_transaction:     &model.Transaction{WalletID: 1, Type: "withdraw", Amount: 10, Category: "food", RelatedWalletID: nil},
			expected_transaction: &model.Transaction{WalletID: 1, Type: "withdraw", Amount: 10, Category: "food", RelatedWalletID: nil},
			expected_err:         nil,
			get_by_id_calls:      1,
			update_calls:         1,
			create_trans_calls:   1,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_by_id_return, Update_calls: usecase.update_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Create_transaction_calls: usecase.create_trans_return, Budget_err: repository.ErrRecordNotFound}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the withdraw service.
			response, err := service.Withdraw_process(usecase.input_request, usecase.input_user)
			//check the response.
			if !errors.Is(err, usecase.expected_err) {
				t.Errorf("Response: expected error %v, got %v", usecase.expected_err, err)
			}
			//check the repo calls.
			if mock_wallet_repo.Get_by_id_count != usecase.get_by_id_calls {
				t.Errorf("Behaviour: expected get_wallet_by_id calls %d, got %d", usecase.get_by_id_calls, mock_wallet_repo.Get_by_id_count)
			}
			if mock_wallet_repo.Update_count != usecase.update_calls {
				t.Errorf("Behaviour: expected update_wallet_balance calls %d, got %d", usecase.update_calls, mock_wallet_repo.Update_count)
			}
			if mock_trans_repo.Create_transaction_count != usecase.create_trans_calls {
				t.Errorf("Behaviour: expected create_transaction calls %d, got %d", usecase.create_trans_calls, mock_trans_repo.Create_transaction_count)
			}
			//check the repo input.
			if usecase.update_calls > 0 && len(mock_wallet_repo.Update_calls) > 0 {
				//check the new balance.
				received_balance := mock_wallet_repo.Update_calls[0].Received_balance
				if received_balance != usecase.repo_balance {
					t.Errorf("Logic: expected new wallet balance in repo %d got %d", usecase.repo_balance, received_balance)
				}
			}
			if usecase.create_trans_calls > 0 && len(mock_trans_repo.Create_transaction_calls) > 0 {
				//check the new balance.
				received_transaction := mock_trans_repo.Create_transaction_calls[0].Received_transaction
				//compare received with expected.
				utils.Compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_transaction)
			}
			//comare the output transaction to expected.
			if response == nil && usecase.expected_transaction != nil {
				t.Errorf("Response: expected not nil transaction got a nil response struct")
			} else if response != nil && usecase.expected_transaction == nil {
				t.Errorf("Response: expected nil response struct got not nil response struct")
			} else if response != nil {
				utils.Compare_transaction_structs(t, "Response", response.Transaction, usecase.expected_transaction)
			}
		})
	}
}

func Test_transfer_process(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name                  string
		input_request         model.Transfer_request
		input_user            *model.User_claims
		get_by_id_return      []mock.Data
		wanted_user_return    *model.User
		wanted_get_user_err   error
		repo_username         string
		update_wallet_return  []mock.Data
		repo_sender_balance   int
		repo_sender_id        int
		repo_receiver_balance int
		repo_receiver_id      int
		create_trans_return   []mock.Trans_data
		repo_out_transaction  *model.Transaction
		repo_in_transaction   *model.Transaction
		expected_transaction  *model.Transaction
		expected_err          error
		get_by_id_calls       int
		update_calls          int
		create_trans_calls    int
		call_get_user         bool
	}
	//set up a pointer to integer to use it in success case.
	sender_wallet := 1
	receiver_wallet := 2
	//create a table of usecase.
	tests := []usecase{
		{
			name:                 "Error: Invalid transaction amount",
			input_request:        model.Transfer_request{Amount: -10, Category: "		"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrInvalidAmount,
			get_by_id_calls:      0,
			update_calls:         0,
			create_trans_calls:   0,
			call_get_user:        false,
		}, {
			name:                 "Error: Empty category",
			input_request:        model.Transfer_request{Amount: 1, Category: "		"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrEmptyCategory,
			get_by_id_calls:      0,
			update_calls:         0,
			create_trans_calls:   0,
			call_get_user:        false,
		}, {
			name:                 "Error: Empty receiver",
			input_request:        model.Transfer_request{Amount: 1, Category: "	food ", ToUsername: "	"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{},
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrEmptyReceiver,
			get_by_id_calls:      0,
			update_calls:         0,
			create_trans_calls:   0,
			call_get_user:        false,
		}, {
			name:                 "Error: sender Wallet not found",
			input_request:        model.Transfer_request{Amount: 1, Category: "	FOOd	", ToUsername: "	MoHamed"},
			input_user:           &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			repo_sender_id:       1,
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrWalletNotFound,
			get_by_id_calls:      1,
			update_calls:         0,
			create_trans_calls:   0,
			call_get_user:        false,
		}, {
			name:          "Error: Insufficient balance",
			input_request: model.Transfer_request{Amount: 20, Category: "	FOOd	", ToUsername: "	MoHamed"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 10, ID: sender_wallet, UserID: 1},
				Wanted_err: nil}},
			repo_sender_id:       1,
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrInSufficient,
			get_by_id_calls:      1,
			update_calls:         0,
			create_trans_calls:   0,
			call_get_user:        false,
		}, {
			name:          "Error: Receiver not found",
			input_request: model.Transfer_request{Amount: 20, Category: "	FOOd	", ToUsername: "	MohAmed	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: sender_wallet, UserID: 1},
				Wanted_err: nil}},
			wanted_user_return:   nil,
			wanted_get_user_err:  repository.ErrUserNotFound,
			repo_username:        "mohamed",
			repo_sender_id:       1,
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrReceiverNotFound,
			get_by_id_calls:      1,
			update_calls:         0,
			create_trans_calls:   0,
			call_get_user:        true,
		}, {
			name:          "Error: Sender is the same as the receiver",
			input_request: model.Transfer_request{Amount: 30, Category: "	FOOd	", ToUsername: "	ALi "},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: sender_wallet, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1}, Wanted_err: nil}},
			wanted_user_return:   &model.User{Username: "ali", ID: 1, Role: "user"},
			wanted_get_user_err:  nil,
			repo_username:        "ali",
			repo_sender_id:       1,
			repo_receiver_id:     1,
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			repo_out_transaction: nil,
			expected_transaction: nil,
			expected_err:         ErrInvalidTransfer,
			get_by_id_calls:      2,
			update_calls:         0,
			create_trans_calls:   0,
			call_get_user:        true,
		}, {
			name:          "Error: Server error with update balance",
			input_request: model.Transfer_request{Amount: 30, Category: "	FOOd	", ToUsername: "	Mohamed "},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: sender_wallet, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 10, ID: receiver_wallet, UserID: 2}, Wanted_err: nil}},
			wanted_user_return:   &model.User{Username: "mohamed", ID: 2, Role: "user"},
			wanted_get_user_err:  nil,
			repo_username:        "mohamed",
			update_wallet_return: []mock.Data{{Wanted_err: errors.New("Database connection failure")}},
			repo_sender_balance:  70,
			repo_sender_id:       1,
			repo_receiver_id:     2,
			create_trans_return:  []mock.Trans_data{},
			repo_out_transaction: nil,
			expected_transaction: nil,
			expected_err:         ErrServerError,
			get_by_id_calls:      2,
			update_calls:         1,
			create_trans_calls:   0,
			call_get_user:        true,
		}, {
			name:          "Error: Server error with create transaction record",
			input_request: model.Transfer_request{Amount: 10, Category: "	FOOd	", ToUsername: "	Mohamed "},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 50, ID: sender_wallet, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 10, ID: receiver_wallet, UserID: 2}, Wanted_err: nil}},
			wanted_user_return:    &model.User{Username: "mohamed", ID: receiver_wallet, Role: "user"},
			wanted_get_user_err:   nil,
			repo_username:         "mohamed",
			update_wallet_return:  []mock.Data{{Wanted_err: nil}, {Wanted_err: nil}},
			repo_sender_balance:   40,
			repo_sender_id:        1,
			repo_receiver_balance: 20,
			repo_receiver_id:      2,
			create_trans_return:   []mock.Trans_data{{Wanted_err: errors.New("Database connection failure")}},
			repo_out_transaction:  &model.Transaction{WalletID: sender_wallet, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: &receiver_wallet},
			expected_transaction:  nil,
			expected_err:          ErrServerError,
			get_by_id_calls:       2,
			update_calls:          2,
			create_trans_calls:    1,
			call_get_user:         true,
		}, {
			name:          "Success: Transaction created",
			input_request: model.Transfer_request{Amount: 10, Category: "	FOOd	", ToUsername: "	Mohamed "},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 30, ID: sender_wallet, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 10, ID: receiver_wallet, UserID: 2}, Wanted_err: nil}},
			wanted_user_return:    &model.User{Username: "mohamed", ID: 2, Role: "user"},
			wanted_get_user_err:   nil,
			repo_username:         "mohamed",
			update_wallet_return:  []mock.Data{{Wanted_err: nil}, {Wanted_err: nil}},
			repo_sender_balance:   20,
			repo_sender_id:        1,
			repo_receiver_balance: 20,
			repo_receiver_id:      2,
			create_trans_return:   []mock.Trans_data{{Wanted_err: nil}, {Wanted_err: nil}},
			repo_out_transaction:  &model.Transaction{WalletID: sender_wallet, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: &receiver_wallet},
			repo_in_transaction:   &model.Transaction{WalletID: receiver_wallet, Type: "transfer_in", Amount: 10, Category: "food", RelatedWalletID: &sender_wallet},
			expected_transaction:  &model.Transaction{WalletID: sender_wallet, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: &receiver_wallet},
			expected_err:          nil,
			get_by_id_calls:       2,
			update_calls:          2,
			create_trans_calls:    2,
			call_get_user:         true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_by_id_return, Update_calls: usecase.update_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Create_transaction_calls: usecase.create_trans_return, Budget_err: repository.ErrRecordNotFound}
			mock_user_repo := &mock.Mock_user_repo{Wanted_user: usecase.wanted_user_return, Wanted_err: usecase.wanted_get_user_err}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the transfer service.
			response, err := service.Transfer_process(usecase.input_request, usecase.input_user)
			//check the response.
			if !errors.Is(err, usecase.expected_err) {
				t.Errorf("Response: expected error %v, got %v", usecase.expected_err, err)
			}
			//check the repo calls.
			if mock_wallet_repo.Get_by_id_count != usecase.get_by_id_calls {
				t.Errorf("Behaviour: expected get_wallet_by_id calls %d, got %d", usecase.get_by_id_calls, mock_wallet_repo.Get_by_id_count)
			}
			if mock_wallet_repo.Update_count != usecase.update_calls {
				t.Errorf("Behaviour: expected update_wallet_balance calls %d, got %d", usecase.update_calls, mock_wallet_repo.Update_count)
			}
			if mock_trans_repo.Create_transaction_count != usecase.create_trans_calls {
				t.Errorf("Behaviour: expected create_transaction calls %d, got %d", usecase.create_trans_calls, mock_trans_repo.Create_transaction_count)
			}
			if mock_user_repo.Was_called != usecase.call_get_user {
				t.Errorf("Behaviour: expected Get_user_by_name call %t, got %t", usecase.call_get_user, mock_user_repo.Was_called)
			}
			//check the repo input.
			if mock_user_repo.Was_called == true && usecase.call_get_user == true {
				//check the passed username.
				if mock_user_repo.Received_name != usecase.repo_username {
					t.Errorf("Logic: expected username in get_user_by_name call %s got %s", usecase.repo_username, mock_user_repo.Received_name)
				}
			}
			if mock_wallet_repo.Get_by_id_count >= 1 && usecase.get_by_id_calls >= 1 {
				//check the sender_id.
				received_id := mock_wallet_repo.Get_by_id_calls[0].Received_id
				if received_id != usecase.repo_sender_id {
					t.Errorf("Logic: expected sender ID passed to get_wallet_by_userID %d got %d", usecase.repo_sender_id, received_id)
				}
			}
			if mock_wallet_repo.Get_by_id_count >= 2 && usecase.get_by_id_calls >= 2 {
				//check the sender_id.
				received_id := mock_wallet_repo.Get_by_id_calls[1].Received_id
				if received_id != usecase.repo_receiver_id {
					t.Errorf("Logic: expected receiver ID passed to get_wallet_by_userID %d got %d", usecase.repo_receiver_id, received_id)
				}
			}
			if usecase.update_calls >= 1 && len(mock_wallet_repo.Update_calls) >= 1 {
				//check the new balance.
				received_balance := mock_wallet_repo.Update_calls[0].Received_balance
				received_id := mock_wallet_repo.Update_calls[0].Received_id
				if received_balance != usecase.repo_sender_balance {
					t.Errorf("Logic: expected new sender wallet balance in repo %d got %d", usecase.repo_sender_balance, received_balance)
				}
				if received_id != usecase.repo_sender_id {
					t.Errorf("Logic: expected sender wallet ID in repo %d got %d", usecase.repo_sender_id, received_id)
				}
			}
			//check the second call input.
			if usecase.update_calls >= 2 && len(mock_wallet_repo.Update_calls) >= 2 {
				//check the new balance.
				received_balance := mock_wallet_repo.Update_calls[1].Received_balance
				received_id := mock_wallet_repo.Update_calls[1].Received_id
				if received_balance != usecase.repo_receiver_balance {
					t.Errorf("Logic: expected new receiver wallet balance in repo %d got %d", usecase.repo_receiver_balance, received_balance)
				}
				if received_id != usecase.repo_receiver_id {
					t.Errorf("Logic: expected receiver wallet ID in repo %d got %d", usecase.repo_receiver_id, received_id)
				}
			}
			if usecase.create_trans_calls >= 1 && len(mock_trans_repo.Create_transaction_calls) >= 1 {
				//check the new balance.
				received_transaction := mock_trans_repo.Create_transaction_calls[0].Received_transaction
				//compare received with expected.
				utils.Compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_out_transaction)
			}
			//check the second call input.
			if usecase.create_trans_calls >= 2 && len(mock_trans_repo.Create_transaction_calls) >= 2 {
				//check the new balance.
				received_transaction := mock_trans_repo.Create_transaction_calls[1].Received_transaction
				//compare received with expected.
				utils.Compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_in_transaction)
			}
			//comare the output transaction to expected.
			if response == nil && usecase.expected_transaction != nil {
				t.Errorf("Response: expected not nil transaction got a nil response struct")
			} else if response != nil && usecase.expected_transaction == nil {
				t.Errorf("Response: expected nil response struct got not nil response struct")
			} else if response != nil {
				utils.Compare_transaction_structs(t, "Response", response.Transaction, usecase.expected_transaction)
			}
		})
	}
}
