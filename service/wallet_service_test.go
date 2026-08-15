package service

import (
	"Wallet-App/mock"
	"Wallet-App/model"
	"Wallet-App/repository"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Setup_testing_db() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to in-memory test database")
	}
	return db
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
		repo_receiver_balance int
		create_trans_return   []mock.Trans_data
		repo_out_transaction  *model.Transaction
		repo_in_transaction   *model.Transaction
		expected_transaction  *model.Transaction
		expected_err          error
		get_by_id_calls       int
		update_calls          int
		create_trans_calls    int
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
		}, {
			name:                 "Error: sender Wallet not found",
			input_request:        model.Transfer_request{Amount: 1, Category: "	FOOd	", ToUsername: "	MoHamed"},
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
			input_request: model.Transfer_request{Amount: 20, Category: "	FOOd	", ToUsername: "	MoHamed"},
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
			name:          "Error: Receiver not found",
			input_request: model.Transfer_request{Amount: 20, Category: "	FOOd	", ToUsername: "	MohAmed	"},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1},
				Wanted_err: nil}},
			wanted_user_return:   nil,
			wanted_get_user_err:  repository.ErrUserNotFound,
			repo_username:        "mohamed",
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			expected_transaction: nil,
			expected_err:         ErrReceiverNotFound,
			get_by_id_calls:      1,
			update_calls:         0,
			create_trans_calls:   0,
		}, {
			name:          "Error: Sender is the same as the receiver",
			input_request: model.Transfer_request{Amount: 30, Category: "	FOOd	", ToUsername: "	ALi "},
			input_user:    &model.User_claims{Username: "ali", UserID: 1, Role: "user"},
			get_by_id_return: []mock.Data{{Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1},
				Wanted_err: nil}, {Wanted_wallet: &model.Wallet{Balance: 100, ID: 1, UserID: 1}, Wanted_err: nil}},
			wanted_user_return:   &model.User{Username: "ali", ID: 1, Role: "user"},
			wanted_get_user_err:  nil,
			repo_username:        "ali",
			update_wallet_return: []mock.Data{},
			create_trans_return:  []mock.Trans_data{},
			repo_out_transaction: nil,
			expected_transaction: nil,
			expected_err:         ErrInvalidTransfer,
			get_by_id_calls:      2,
			update_calls:         0,
			create_trans_calls:   0,
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
			create_trans_return:  []mock.Trans_data{},
			repo_out_transaction: nil,
			expected_transaction: nil,
			expected_err:         ErrServerError,
			get_by_id_calls:      2,
			update_calls:         1,
			create_trans_calls:   0,
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
			repo_receiver_balance: 20,
			create_trans_return:   []mock.Trans_data{{Wanted_err: errors.New("Database connection failure")}},
			repo_out_transaction:  &model.Transaction{WalletID: 1, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: ptr},
			expected_transaction:  nil,
			expected_err:          ErrServerError,
			get_by_id_calls:       2,
			update_calls:          2,
			create_trans_calls:    1,
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
			repo_receiver_balance: 20,
			create_trans_return:   []mock.Trans_data{{Wanted_err: nil}, {Wanted_err: nil}},
			repo_out_transaction:  &model.Transaction{WalletID: 1, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: ptr},
			repo_in_transaction:   &model.Transaction{WalletID: 2, Type: "transfer_in", Amount: 10, Category: "food", RelatedWalletID: ptr},
			expected_transaction:  &model.Transaction{WalletID: 1, Type: "transfer_out", Amount: 10, Category: "food", RelatedWalletID: ptr},
			expected_err:          nil,
			get_by_id_calls:       2,
			update_calls:          2,
			create_trans_calls:    2,
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
			//check the repo input.
			if usecase.update_calls >= 1 && len(mock_wallet_repo.Update_calls) >= 1 {
				//check the new balance.
				received_balance := mock_wallet_repo.Update_calls[0].Received_balance
				if received_balance != usecase.repo_sender_balance {
					t.Errorf("Logic: expected new sender wallet balance in repo %d got %d", usecase.repo_sender_balance, received_balance)
				}
			}
			//check the second call input.
			if usecase.update_calls >= 2 && len(mock_wallet_repo.Update_calls) >= 2 {
				//check the new balance.
				received_balance := mock_wallet_repo.Update_calls[1].Received_balance
				if received_balance != usecase.repo_receiver_balance {
					t.Errorf("Logic: expected new receiver wallet balance in repo %d got %d", usecase.repo_receiver_balance, received_balance)
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
