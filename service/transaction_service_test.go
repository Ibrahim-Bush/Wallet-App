package service

import (
	"Wallet-App/mock"
	"Wallet-App/model"
	"Wallet-App/repository"
	"errors"
	"testing"
	"time"
)

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
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_list: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			//pass mocks to the service.
			service := Init_transaction_service(mock_trans_repo, mock_wallet_repo)
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
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_list: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			//pass mocks to the service.
			service := Init_transaction_service(mock_trans_repo, mock_wallet_repo)
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
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_list: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			//pass mocks to the service.
			service := Init_transaction_service(mock_trans_repo, mock_wallet_repo)
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
			mock_wallet_repo := &mock.Mock_wallet_repo{Get_by_id_calls: usecase.get_wallet_return}
			mock_trans_repo := &mock.Mock_transaction_repo{Wanted_summary: usecase.repo_transactions, Wanted_err: usecase.get_records_err}
			//pass mocks to the service.
			service := Init_transaction_service(mock_trans_repo, mock_wallet_repo)
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
