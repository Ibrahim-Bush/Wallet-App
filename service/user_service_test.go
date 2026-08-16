package service

import (
	"Wallet-App/mock"
	"Wallet-App/model"
	"Wallet-App/repository"
	"Wallet-App/utils"
	"errors"
	"strings"
	"testing"
)

func Test_create_user(t *testing.T) {

	//define a struct for usecases.
	type usecase struct {
		name               string
		input_data         model.Auth_request
		repo_user          *model.User
		create_user_err    error
		repo_wallet        *model.Wallet
		create_wallet_err  error
		expected_err       error
		expected_user      *model.User
		expected_wallet    *model.Wallet
		create_user_call   bool
		create_wallet_call bool
	}
	//create a table of usecases.
	tests := []usecase{
		{
			name:               "Error: Empty username",
			input_data:         model.Auth_request{Username: "	", Password: ""},
			repo_user:          nil,
			create_user_err:    nil,
			repo_wallet:        nil,
			create_wallet_err:  nil,
			expected_err:       ErrEmptyName,
			expected_user:      nil,
			expected_wallet:    nil,
			create_user_call:   false,
			create_wallet_call: false,
		}, {
			name:               "Error: Empty password",
			input_data:         model.Auth_request{Username: " ahmed ", Password: "		"},
			repo_user:          nil,
			create_user_err:    nil,
			repo_wallet:        nil,
			create_wallet_err:  nil,
			expected_err:       ErrEmptyPassword,
			expected_user:      nil,
			expected_wallet:    nil,
			create_user_call:   false,
			create_wallet_call: false,
		}, {
			name:               "Error: Invalid password",
			input_data:         model.Auth_request{Username: " ahmed ", Password: strings.Repeat("a", 200)},
			repo_user:          nil,
			create_user_err:    nil,
			repo_wallet:        nil,
			create_wallet_err:  nil,
			expected_err:       ErrInvalidPassword,
			expected_user:      nil,
			expected_wallet:    nil,
			create_user_call:   false,
			create_wallet_call: false,
		}, {
			name:               "Error: Duplicate username",
			input_data:         model.Auth_request{Username: "	AhMeD	", Password: "	Admin	"},
			repo_user:          &model.User{Username: "ahmed", Password: "Admin", Role: "user"},
			create_user_err:    repository.ErrDuplicatedName,
			repo_wallet:        nil,
			create_wallet_err:  nil,
			expected_err:       ErrDuplicatedName,
			expected_user:      nil,
			expected_wallet:    nil,
			create_user_call:   true,
			create_wallet_call: false,
		}, {
			name:               "Error: Server error with create user",
			input_data:         model.Auth_request{Username: "	AhMeD	", Password: "	Admin	"},
			repo_user:          &model.User{Username: "ahmed", Password: "Admin", Role: "user"},
			create_user_err:    errors.New("database connection failure"),
			repo_wallet:        nil,
			create_wallet_err:  nil,
			expected_err:       ErrServerError,
			expected_user:      nil,
			expected_wallet:    nil,
			create_user_call:   true,
			create_wallet_call: false,
		}, {
			name:               "Error: Duplicate wallet",
			input_data:         model.Auth_request{Username: "	AhMeD	", Password: "	Admin	"},
			repo_user:          &model.User{Username: "ahmed", Password: "Admin", Role: "user"},
			create_user_err:    nil,
			repo_wallet:        &model.Wallet{Balance: 0},
			create_wallet_err:  repository.ErrDuplicatedWallet,
			expected_err:       ErrDuplicatedWallet,
			expected_user:      nil,
			expected_wallet:    nil,
			create_user_call:   true,
			create_wallet_call: true,
		}, {
			name:               "Error: Server error with create wallet",
			input_data:         model.Auth_request{Username: "	AhMeD	", Password: "	Admin	"},
			repo_user:          &model.User{Username: "ahmed", Password: "Admin", Role: "user"},
			create_user_err:    nil,
			repo_wallet:        &model.Wallet{Balance: 0},
			create_wallet_err:  errors.New("Database connection failure"),
			expected_err:       ErrServerError,
			expected_user:      nil,
			expected_wallet:    nil,
			create_user_call:   true,
			create_wallet_call: true,
		}, {
			name:               "Success: user created with wallet",
			input_data:         model.Auth_request{Username: "	AhMeD	", Password: "	Admin	"},
			repo_user:          &model.User{Username: "ahmed", Password: "Admin", Role: "user"},
			create_user_err:    nil,
			repo_wallet:        &model.Wallet{Balance: 0},
			create_wallet_err:  nil,
			expected_err:       nil,
			expected_user:      &model.User{Username: "ahmed", Password: "Admin", Role: "user"},
			expected_wallet:    &model.Wallet{Balance: 0},
			create_user_call:   true,
			create_wallet_call: true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//create a new mock repo.
			mock_user_repo := &mock.Mock_user_repo{Wanted_err: usecase.create_user_err}
			mock_wallet_repo := &mock.Mock_wallet_repo{Wanted_err: usecase.create_wallet_err}
			testDB := Setup_testing_db()
			//pass mock to the service.
			service := Init_user_service(mock_user_repo, mock_wallet_repo, testDB)
			//call method with usecases.
			user, wallet, err := service.Create_user(usecase.input_data)
			//check the repository call: first check the call status.
			if mock_user_repo.Was_called != usecase.create_user_call {
				t.Errorf("Behaviour: Expected create_user method call %t got %t", usecase.create_user_call, mock_user_repo.Was_called)
			}
			if mock_wallet_repo.Was_called != usecase.create_wallet_call {
				t.Errorf("Behaviour: Expected create_wallet method call %t got %t", usecase.create_wallet_call, mock_wallet_repo.Was_called)
			}
			//compare expected repo struct to the received struct.
			compare_user_structs(t, "Repo input", mock_user_repo.Received_user, usecase.repo_user)
			//compare wallet repo input.
			compare_wallet_structs(t, "Repo input", mock_wallet_repo.Received_wallet, usecase.repo_wallet)
			//check responeses.
			if err != usecase.expected_err {
				t.Errorf("Response: Expected error %v got %v", usecase.expected_err, err)
			}
			//check fields in response.
			compare_user_structs(t, "Service response", user, usecase.expected_user)
			//check the response wallet.
			compare_wallet_structs(t, "Service response", wallet, usecase.expected_wallet)
		})
	}
}

func compare_user_structs(t *testing.T, stage string, received *model.User, expected *model.User) {
	//first compare pointers.
	if received == nil || expected == nil {
		if received == nil && expected != nil {
			t.Errorf("Boundary (%s): expected struct %+v got nil pointer", stage, expected)
		} else if received != nil && expected == nil {
			t.Errorf("Boundary (%s): expected nil struct got %+v", stage, received)
		}
	} else {
		//compare received struct fields to the expected struct fields.
		if received.Username != expected.Username {
			t.Errorf("Logic (%s): expected username %s got %s", stage, expected.Username, received.Username)
		}
		if received.Password == expected.Password {
			t.Errorf("Logic (%s): expected hashed Password got password as plain text", stage)
		} else if utils.Compare_password(expected.Password, received.Password) == false {
			t.Errorf("Logic (%s): Hashed password does not match original password", stage)
		}
		if received.Role != expected.Role {
			t.Errorf("Logic (%s): expected role %s got %s", stage, expected.Role, received.Role)
		}
	}
}

func compare_wallet_structs(t *testing.T, stage string, received *model.Wallet, expected *model.Wallet) {
	//first compare pointers.
	if received == nil || expected == nil {
		if received == nil && expected != nil {
			t.Errorf("Boundary (%s): expected struct %+v got nil pointer", stage, expected)
		} else if received != nil && expected == nil {
			t.Errorf("Boundary (%s): expected nil struct got %+v", stage, received)
		}
	} else {
		if received.Balance != expected.Balance {
			t.Errorf("Logic (%s): expected Balance %d got %d", stage, expected.Balance, received.Balance)
		}
	}
}

func Test_login_user(t *testing.T) {
	//define a struct for usecases.
	type usecase struct {
		name           string
		input_data     model.Auth_request
		repo_name      string
		repo_err       error
		repo_user      *model.User
		expected_token string
		expected_err   error
		call_repo      bool
	}
	//create a hashed password for success case.
	hashed_admin, err := utils.Hash_password("admin")
	if err != nil {
		t.Fatalf("Error from testing func: could not make a hashed password for success case")
	}
	//create a table of usecases.
	tests := []usecase{
		{
			name:           "Error: User not found",
			input_data:     model.Auth_request{Username: "  AhMEd	", Password: "		"},
			repo_name:      "ahmed",
			repo_err:       repository.ErrUserNotFound,
			repo_user:      nil,
			expected_token: "",
			expected_err:   ErrUserNotFound,
			call_repo:      true,
		}, {
			name:           "Error: Server error",
			input_data:     model.Auth_request{Username: "  AhMEd	", Password: "	abc	"},
			repo_name:      "ahmed",
			repo_err:       errors.New("Database connection failure"),
			repo_user:      nil,
			expected_token: "",
			expected_err:   ErrServerError,
			call_repo:      true,
		}, {
			name:           "Error: Wrong password",
			input_data:     model.Auth_request{Username: "  AhMEd	", Password: "	abc	"},
			repo_name:      "ahmed",
			repo_err:       nil,
			repo_user:      &model.User{Username: "ahmed", Password: "admin"},
			expected_token: "",
			expected_err:   ErrWrongPassword,
			call_repo:      true,
		}, {
			name:           "Error: Wrong password",
			input_data:     model.Auth_request{Username: "  AhMEd	", Password: "	abc	"},
			repo_name:      "ahmed",
			repo_err:       nil,
			repo_user:      &model.User{Username: "ahmed", Password: "admin"},
			expected_token: "",
			expected_err:   ErrWrongPassword,
			call_repo:      true,
		}, {
			name:           "Success: get token",
			input_data:     model.Auth_request{Username: "  AhMEd	", Password: "	admin	"},
			repo_name:      "ahmed",
			repo_err:       nil,
			repo_user:      &model.User{Username: "ahmed", Password: hashed_admin, ID: 1, Role: "user"},
			expected_token: "not null",
			expected_err:   nil,
			call_repo:      true,
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//create new mock repo with case data.
			mock_repo := &mock.Mock_user_repo{
				Wanted_err:  usecase.repo_err,
				Wanted_user: usecase.repo_user,
			}
			mock_wallet_repo := &mock.Mock_wallet_repo{}
			testDB := Setup_testing_db()
			//pass mock to the service.
			service := Init_user_service(mock_repo, mock_wallet_repo, testDB)
			//call method with usecase.
			token, err := service.Login_user(usecase.input_data)
			//check the repo input: first check the call status.
			if mock_repo.Was_called != usecase.call_repo {
				t.Errorf("Behaviour: Expected repository call %t got %t", usecase.call_repo, mock_repo.Was_called)
			}
			//check the fields in the repo call.
			if mock_repo.Received_name != usecase.repo_name {
				t.Errorf("Logic (Repo input): Expected username %s got %s", usecase.repo_name, mock_repo.Received_name)
			}
			//check the service response.
			if err != usecase.expected_err {
				t.Errorf("Response: Expected error %v got %v", usecase.expected_err, err)
			}
			if usecase.expected_token == "" && token != usecase.expected_token {
				t.Errorf("Response: Expected token %s got %s", usecase.expected_token, token)
			} else if usecase.expected_token != "" {
				//verify the token in response.
				claims, err := utils.Verify_token(token)
				if claims != nil && err == nil {
					if claims.Username != usecase.repo_user.Username {
						t.Errorf("Response: Expected username in token %s got %s", usecase.repo_user.Username, claims.Username)
					}
					if claims.UserID != usecase.repo_user.ID {
						t.Errorf("Response: Expected userID in token %d got %d", usecase.repo_user.ID, claims.UserID)
					}
					if claims.Role != usecase.repo_user.Role {
						t.Errorf("Response: Expected Role in token %s got %s", usecase.repo_user.Role, claims.Role)
					}
				}
			}
		})
	}
}
