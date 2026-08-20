package service

import (
	"Wallet-App/mock"
	"Wallet-App/model"
	"Wallet-App/repository"
	"errors"
	"testing"
	"time"

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
				compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_transaction)
			}
			//comare the output transaction to expected.
			compare_transaction_structs(t, "Response", transaction, usecase.expected_transaction)
		})
	}
}

func compare_transaction_structs(t *testing.T, stage string, received *model.Transaction, expected *model.Transaction) {
	//first compare pointers.
	if received == nil || expected == nil {
		if received == nil && expected != nil {
			t.Errorf("Boundary (%s): expected struct %+v got nil pointer", stage, expected)
		} else if received != nil && expected == nil {
			t.Errorf("Boundary (%s): expected nil struct got %+v", stage, received)
		}
	} else {
		//compare received struct fields to the expected struct fields.
		if received.WalletID != expected.WalletID {
			t.Errorf("Logic (%s): expected WalletID %d got %d", stage, expected.WalletID, received.WalletID)
		}
		if received.Type != expected.Type {
			t.Errorf("Logic (%s): expected type %s got %s", stage, expected.Type, received.Type)
		}
		if received.Amount != expected.Amount {
			t.Errorf("Logic (%s): expected Amount %d got %d", stage, expected.Amount, received.Amount)
		}
		if received.Category != expected.Category {
			t.Errorf("Logic (%s): expected category %s got %s", stage, expected.Category, received.Category)
		}
		if expected.RelatedWalletID == nil && received.RelatedWalletID != expected.RelatedWalletID {
			t.Errorf("Logic (%s): expected RelatedWalletID %v got %v", stage, expected.RelatedWalletID, received.RelatedWalletID)
		} else if expected.RelatedWalletID != nil && received.RelatedWalletID == nil {
			t.Errorf("Logic (%s): expected not nil RelatedWalletID got %v", stage, received.RelatedWalletID)
		}
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
			mock_trans_repo := &mock.Mock_transaction_repo{Create_transaction_calls: usecase.create_trans_return}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the withdraw service.
			transaction, err := service.Withdraw_process(usecase.input_request, usecase.input_user)
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
				compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_transaction)
			}
			//comare the output transaction to expected.
			compare_transaction_structs(t, "Response", transaction, usecase.expected_transaction)
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
	num := 10
	var ptr *int = &num
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
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 10, ID: 1, UserID: 1},
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
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1},
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
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1},
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
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 10, ID: 2, UserID: 2}, Wanted_err: nil}},
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
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 50, ID: 1, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 10, ID: 2, UserID: 2}, Wanted_err: nil}},
			wanted_user_return:    &model.User{Username: "mohamed", ID: 2, Role: "user"},
			wanted_get_user_err:   nil,
			repo_username:         "mohamed",
			update_wallet_return:  []mock.Data{{Wanted_err: nil}, {Wanted_err: nil}},
			repo_sender_balance:   40,
			repo_sender_id:        1,
			repo_receiver_balance: 20,
			repo_receiver_id:      2,
			create_trans_return:   []mock.Trans_data{{Wanted_err: errors.New("Database connection failure")}},
			repo_out_transaction:  &model.Transaction{WalletID: 1, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: ptr},
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
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 30, ID: 1, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 10, ID: 2, UserID: 2}, Wanted_err: nil}},
			wanted_user_return:    &model.User{Username: "mohamed", ID: 2, Role: "user"},
			wanted_get_user_err:   nil,
			repo_username:         "mohamed",
			update_wallet_return:  []mock.Data{{Wanted_err: nil}, {Wanted_err: nil}},
			repo_sender_balance:   20,
			repo_sender_id:        1,
			repo_receiver_balance: 20,
			repo_receiver_id:      2,
			create_trans_return:   []mock.Trans_data{{Wanted_err: nil}, {Wanted_err: nil}},
			repo_out_transaction:  &model.Transaction{WalletID: 1, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: ptr},
			repo_in_transaction:   &model.Transaction{WalletID: 2, Type: "transfer_in", Amount: 10, Category: "food", RelatedWalletID: ptr},
			expected_transaction:  &model.Transaction{WalletID: 1, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: ptr},
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
			mock_trans_repo := &mock.Mock_transaction_repo{Create_transaction_calls: usecase.create_trans_return}
			mock_user_repo := &mock.Mock_user_repo{Wanted_user: usecase.wanted_user_return, Wanted_err: usecase.wanted_get_user_err}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the transfer service.
			transaction, err := service.Transfer_process(usecase.input_request, usecase.input_user)
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
				compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_out_transaction)
			}
			//check the second call input.
			if usecase.create_trans_calls >= 2 && len(mock_trans_repo.Create_transaction_calls) >= 2 {
				//check the new balance.
				received_transaction := mock_trans_repo.Create_transaction_calls[1].Received_transaction
				//compare received with expected.
				compare_transaction_structs(t, "Repo input", received_transaction, usecase.repo_in_transaction)
			}
			//comare the output transaction to expected.
			compare_transaction_structs(t, "Response", transaction, usecase.expected_transaction)
		})
	}
}

func Test_get_user_transactions(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name                  string
		input_user            *model.User_claims
		get_wallet_return     []mock.Data
		repo_id               int
		repo_transactions     []model.Transaction
		get_records_err       error
		expected_transactions []model.Transaction
		expected_err          error
		get_wallet_call       int
		get_records_call      bool
	}
	//set a table of usecase.
	tests := []usecase{
		{
			name:                  "Error: Invalid user claims",
			input_user:            nil,
			expected_transactions: nil,
			expected_err:          ErrInvalidUser,
			get_wallet_call:       0,
			get_records_call:      false,
		}, {
			name:                  "Error: Wallet not found",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrWalletNotFound,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get wallet",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: errors.New("Database connection failure")}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get transactions",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_transactions:     nil,
			get_records_err:       errors.New("Database connection failure"),
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      true,
		}, {
			name:                  "Error: Get transactions successfully",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_transactions:     []model.Transaction{{ID: 1}, {ID: 2}, {ID: 3}},
			get_records_err:       nil,
			expected_transactions: []model.Transaction{{ID: 1}, {ID: 2}, {ID: 3}},
			expected_err:          nil,
			get_wallet_call:       1,
			get_records_call:      true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_list: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the get_transactions service.
			list, err := service.Get_user_transactions(usecase.input_user)
			//check the response.
			if err != usecase.expected_err {
				t.Errorf("Response: expected error in response %v got %v", usecase.expected_err, err)
			}
			//check the list in response.
			if usecase.expected_transactions == nil && list != nil {
				t.Errorf("Response: expected nil slice got %+v transactions", list)
			} else if expected := len(usecase.expected_transactions); expected != len(list) {
				t.Errorf("Response: expected slice of length %d got slice of length %d", expected, len(list))
			}
			//check the repo call.
			if mock_wallet_repo.Get_by_id_count != usecase.get_wallet_call {
				t.Errorf("Behaviour: expected get_wallet_by_id calls %d, got %d", usecase.get_wallet_call, mock_wallet_repo.Get_by_id_count)
			}
			if mock_wallet_repo.Get_by_id_count > 0 && usecase.get_wallet_call > 0 {
				//check the passed user id.
				received_id := mock_wallet_repo.Get_by_id_calls[0].Received_id
				if received_id != usecase.repo_id {
					t.Errorf("Logic: expected user ID passed to get_wallet_by_userID %d got %d", usecase.repo_id, received_id)
				}
			}
			if mock_trans_repo.Was_called != usecase.get_records_call {
				t.Errorf("Behaviour: expected repository method call %t got %t", usecase.get_records_call, mock_trans_repo.Was_called)
			}
		})
	}
}

func Test_get_transactions_by_category(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name                  string
		input_user            *model.User_claims
		input_category        string
		get_wallet_return     []mock.Data
		repo_id               int
		repo_transactions     []model.Transaction
		repo_category         string
		get_records_err       error
		expected_transactions []model.Transaction
		expected_err          error
		get_wallet_call       int
		get_records_call      bool
	}
	//set a table of usecase.
	tests := []usecase{
		{
			name:                  "Error: Invalid user claims",
			input_user:            nil,
			expected_transactions: nil,
			expected_err:          ErrInvalidUser,
			get_wallet_call:       0,
			get_records_call:      false,
		},
		{
			name:                  "Error: Empty category",
			input_category:        "		",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			expected_transactions: nil,
			expected_err:          ErrEmptyCategory,
			get_wallet_call:       0,
			get_records_call:      false,
		}, {
			name:                  "Error: Wallet not found",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_category:        "	FoOD	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrWalletNotFound,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get wallet",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_category:        "	FoOD	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: errors.New("Database connection failure")}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get transactions",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_category:        "	FoOD	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_category:         "food",
			repo_transactions:     nil,
			get_records_err:       errors.New("Database connection failure"),
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      true,
		}, {
			name:                  "Error: Get transactions successfully",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_category:        "	WoRK	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_category:         "work",
			repo_transactions:     []model.Transaction{{ID: 1}, {ID: 2}, {ID: 3}},
			get_records_err:       nil,
			expected_transactions: []model.Transaction{{ID: 1}, {ID: 2}, {ID: 3}},
			expected_err:          nil,
			get_wallet_call:       1,
			get_records_call:      true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_list: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the get_transactions_by_category service.
			list, err := service.Get_transactions_by_category(usecase.input_category, usecase.input_user)
			//check the response.
			if err != usecase.expected_err {
				t.Errorf("Response: expected error in response %v got %v", usecase.expected_err, err)
			}
			//check the list in response.
			if usecase.expected_transactions == nil && list != nil {
				t.Errorf("Response: expected nil slice got %+v transactions", list)
			} else if expected := len(usecase.expected_transactions); expected != len(list) {
				t.Errorf("Response: expected slice of length %d got slice of length %d", expected, len(list))
			}
			//check the repo call.
			if mock_wallet_repo.Get_by_id_count != usecase.get_wallet_call {
				t.Errorf("Behaviour: expected get_wallet_by_id calls %d, got %d", usecase.get_wallet_call, mock_wallet_repo.Get_by_id_count)
			}
			if mock_wallet_repo.Get_by_id_count > 0 && usecase.get_wallet_call > 0 {
				//check the passed user id.
				received_id := mock_wallet_repo.Get_by_id_calls[0].Received_id
				if received_id != usecase.repo_id {
					t.Errorf("Logic: expected user ID passed to get_wallet_by_userID %d got %d", usecase.repo_id, received_id)
				}
			}
			if mock_trans_repo.Was_called != usecase.get_records_call {
				t.Errorf("Behaviour: expected repository method call %t got %t", usecase.get_records_call, mock_trans_repo.Was_called)
			}
			if mock_trans_repo.Was_called == true && usecase.get_records_call == true {
				if mock_trans_repo.Received_category != usecase.repo_category {
					t.Errorf("Logic: expected category passed to repo %s got %s", usecase.repo_category, mock_trans_repo.Received_category)
				}
			}
		})
	}
}

func Test_get_transactions_by_date(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name                  string
		input_user            *model.User_claims
		input_start_date      string
		input_end_date        string
		get_wallet_return     []mock.Data
		repo_id               int
		repo_transactions     []model.Transaction
		repo_start_date       time.Time
		repo_end_date         time.Time
		get_records_err       error
		expected_transactions []model.Transaction
		expected_err          error
		get_wallet_call       int
		get_records_call      bool
	}
	//set a table of usecase.
	tests := []usecase{
		{
			name:                  "Error: Invalid user claims",
			input_user:            nil,
			expected_transactions: nil,
			expected_err:          ErrInvalidUser,
			get_wallet_call:       0,
			get_records_call:      false,
		},
		{
			name:                  "Error: Invalid date",
			input_start_date:      "		",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			expected_transactions: nil,
			expected_err:          ErrInvalidDate,
			get_wallet_call:       0,
			get_records_call:      false,
		}, {
			name:                  "Error: Invalid date",
			input_start_date:      "	2026-08-01	",
			input_end_date:        "	 October	",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			expected_transactions: nil,
			expected_err:          ErrInvalidDate,
			get_wallet_call:       0,
			get_records_call:      false,
		}, {
			name:                  "Error: Invalid date range",
			input_start_date:      "	2026-08-01	",
			input_end_date:        "	2022-01-01	",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			expected_transactions: nil,
			expected_err:          ErrInvalidDateRange,
			get_wallet_call:       0,
			get_records_call:      false,
		}, {
			name:                  "Error: Wallet not found",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_start_date:      "	2026-07-01	",
			input_end_date:        "	2026-08-01	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrWalletNotFound,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get wallet",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_start_date:      "	2026-07-01	",
			input_end_date:        "	2026-08-01	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: errors.New("Database connection failure")}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get transactions",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_start_date:      "	2026-07-01	",
			input_end_date:        "	2026-08-01	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_start_date:       time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			repo_end_date:         time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			repo_transactions:     nil,
			get_records_err:       errors.New("Database connection failure"),
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      true,
		}, {
			name:                  "Error: Get transactions successfully",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			input_start_date:      "	2026-07-01	",
			input_end_date:        "	2026-08-01	",
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_start_date:       time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC),
			repo_end_date:         time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC),
			repo_transactions:     []model.Transaction{{ID: 1}, {ID: 2}, {ID: 3}},
			get_records_err:       nil,
			expected_transactions: []model.Transaction{{ID: 1}, {ID: 2}, {ID: 3}},
			expected_err:          nil,
			get_wallet_call:       1,
			get_records_call:      true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_list: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the get transactions by date service.
			list, err := service.Get_transactions_by_date(usecase.input_start_date, usecase.input_end_date, usecase.input_user)
			//check the response.
			if err != usecase.expected_err {
				t.Errorf("Response: expected error in response %v got %v", usecase.expected_err, err)
			}
			//check the list in response.
			if usecase.expected_transactions == nil && list != nil {
				t.Errorf("Response: expected nil slice got %+v transactions", list)
			} else if expected := len(usecase.expected_transactions); expected != len(list) {
				t.Errorf("Response: expected slice of length %d got slice of length %d", expected, len(list))
			}
			//check the repo call.
			if mock_wallet_repo.Get_by_id_count != usecase.get_wallet_call {
				t.Errorf("Behaviour: expected get_wallet_by_id calls %d, got %d", usecase.get_wallet_call, mock_wallet_repo.Get_by_id_count)
			}
			if mock_wallet_repo.Get_by_id_count > 0 && usecase.get_wallet_call > 0 {
				//check the passed user id.
				received_id := mock_wallet_repo.Get_by_id_calls[0].Received_id
				if received_id != usecase.repo_id {
					t.Errorf("Logic: expected user ID passed to get_wallet_by_userID %d got %d", usecase.repo_id, received_id)
				}
			}
			if mock_trans_repo.Was_called != usecase.get_records_call {
				t.Errorf("Behaviour: expected repository method call %t got %t", usecase.get_records_call, mock_trans_repo.Was_called)
			}
			if mock_trans_repo.Was_called == true && usecase.get_records_call == true {
				if !mock_trans_repo.Received_start_date.Equal(usecase.repo_start_date) {
					t.Errorf("Logic: expected start date passed to repo %v got %v", usecase.repo_start_date, mock_trans_repo.Received_start_date)
				}
				if !mock_trans_repo.Received_end_date.Equal(usecase.repo_end_date) {
					t.Errorf("Logic: expected end date passed to repo %v got %v", usecase.repo_end_date, mock_trans_repo.Received_end_date)
				}
			}
		})
	}
}

func Test_get_transactions_summary(t *testing.T) {
	//define a struct for usecase.
	type usecase struct {
		name                  string
		input_user            *model.User_claims
		get_wallet_return     []mock.Data
		repo_id               int
		repo_transactions     []model.Transaction_summary
		repo_month            time.Time
		get_records_err       error
		expected_transactions []model.Transaction_summary
		expected_err          error
		get_wallet_call       int
		get_records_call      bool
	}
	//define a variable for the current time.
	var current_time = time.Now().UTC()
	//set a table of usecases.
	tests := []usecase{
		{
			name:                  "Error: Invalid user",
			input_user:            nil,
			expected_transactions: nil,
			expected_err:          ErrInvalidUser,
			get_wallet_call:       0,
			get_records_call:      false,
		}, {
			name:                  "Error: Wallet not found",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: repository.ErrWalletNotFound}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrWalletNotFound,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get wallet",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: nil, Wanted_err: errors.New("Database connection failure")}},
			repo_id:               1,
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      false,
		}, {
			name:                  "Error: Server error with get summary",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_transactions:     nil,
			repo_month:            time.Date(current_time.Year(), current_time.Month(), 1, 0, 0, 0, 0, time.UTC),
			get_records_err:       errors.New("Database connection failure"),
			expected_transactions: nil,
			expected_err:          ErrServerError,
			get_wallet_call:       1,
			get_records_call:      true,
		}, {
			name:                  "Error: Get transactions summary successfully",
			input_user:            &model.User_claims{UserID: 1, Role: "user"},
			get_wallet_return:     []mock.Data{{Wanted_wallet: &model.Wallet{ID: 2}, Wanted_err: nil}},
			repo_id:               1,
			repo_transactions:     []model.Transaction_summary{{Category: "food", Total: 20}, {Category: "work", Total: 70}},
			repo_month:            time.Date(current_time.Year(), current_time.Month(), 1, 0, 0, 0, 0, time.UTC),
			get_records_err:       nil,
			expected_transactions: []model.Transaction_summary{{Category: "food", Total: 20}, {Category: "work", Total: 70}},
			expected_err:          nil,
			get_wallet_call:       1,
			get_records_call:      true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//setup mocks.
			testDB := Setup_testing_db()
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_summary: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			mock_user_repo := &mock.Mock_user_repo{}
			//pass mocks to the service.
			service := Init_wallet_service(mock_wallet_repo, mock_trans_repo, mock_user_repo, testDB)
			//call the get_summary service.
			summary, err := service.Get_transactions_summary(usecase.input_user)
			//check the response.
			if err != usecase.expected_err {
				t.Errorf("Response: expected error in response %v got %v", usecase.expected_err, err)
			}
			//check the list in response.
			if usecase.expected_transactions == nil && summary != nil {
				t.Errorf("Response: expected nil slice got %+v of transactions summary", summary)
			} else if expected := len(usecase.expected_transactions); expected != len(summary) {
				t.Errorf("Response: expected slice of length %d got slice of length %d", expected, len(summary))
			}
			//check the repo call.
			if mock_wallet_repo.Get_by_id_count != usecase.get_wallet_call {
				t.Errorf("Behaviour: expected get_wallet_by_id calls %d, got %d", usecase.get_wallet_call, mock_wallet_repo.Get_by_id_count)
			}
			if mock_wallet_repo.Get_by_id_count > 0 && usecase.get_wallet_call > 0 {
				//check the passed user id.
				received_id := mock_wallet_repo.Get_by_id_calls[0].Received_id
				if received_id != usecase.repo_id {
					t.Errorf("Logic: expected user ID passed to get_wallet_by_userID %d got %d", usecase.repo_id, received_id)
				}
			}
			if mock_trans_repo.Was_called != usecase.get_records_call {
				t.Errorf("Behaviour: expected repository method call %t got %t", usecase.get_records_call, mock_trans_repo.Was_called)
			}
			if mock_trans_repo.Was_called == true && usecase.get_records_call == true {
				if !mock_trans_repo.Received_month.Equal(usecase.repo_month) {
					t.Errorf("Behaviour: expected month date passed to repo %v got %v", usecase.repo_month, mock_trans_repo.Received_month)
				}
			}
		})
	}
}
