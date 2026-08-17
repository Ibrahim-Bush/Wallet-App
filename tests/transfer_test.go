package test

import (
	"Wallet-App/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

var (
	user1 = model.User{Username: "ali", Password: "admin1"}
	user2 = model.User{Username: "mohamed", Password: "admin2"}
)

func Test_transfer_integration(t *testing.T) {

	//first: reset wallets balance to assure balance after operations.
	Reset_wallets_balance(t, 100)
	//then get a token for users to pass authentication middleware.
	user1_token := Login_test_user(t, user1_Auth)
	user1_wallet := Get_wallet_by_username(t, user1)
	user2_wallet := Get_wallet_by_username(t, user2)
	//define the usecase struct.
	type usecase struct {
		name                          string
		endpoint                      string
		input_request                 model.Transfer_request
		token                         string
		sender_user                   model.User
		receiver_user                 model.User
		expected_status               int
		expected_transaction          model.Transaction
		expected_sender_balance       int
		expected_receiver_balance     int
		expected_sender_transaction   model.Transaction
		expected_receiver_transaction model.Transaction
	}
	//define a table for usecases.
	tests := []usecase{
		{
			name:                          "Transfer - Successful happy path",
			endpoint:                      "/transfer",
			input_request:                 model.Transfer_request{Amount: 50, Category: " 	food	", ToUsername: "	MOhamed	"},
			token:                         user1_token,
			sender_user:                   user1,
			receiver_user:                 user2,
			expected_status:               200,
			expected_transaction:          model.Transaction{WalletID: user1_wallet.ID, Amount: 50, Type: "transfer_out", Category: "food", RelatedWalletID: &user2_wallet.ID},
			expected_sender_balance:       50,
			expected_receiver_balance:     150,
			expected_sender_transaction:   model.Transaction{WalletID: user1_wallet.ID, Amount: 50, Type: "transfer_out", Category: "food", RelatedWalletID: &user2_wallet.ID},
			expected_receiver_transaction: model.Transaction{WalletID: user2_wallet.ID, Amount: 50, Type: "transfer_in", Category: "food", RelatedWalletID: &user1_wallet.ID},
		}, {
			name:                        "withdraw - Insufficient fund",
			endpoint:                    "/withdraw",
			input_request:               model.Transfer_request{Amount: 100, Category: " 	food	"},
			token:                       user1_token,
			sender_user:                 user1,
			receiver_user:               model.User{},
			expected_status:             400,
			expected_transaction:        model.Transaction{},
			expected_sender_balance:     50,
			expected_sender_transaction: model.Transaction{},
		}, {
			name:                        "deposit - Successful process",
			endpoint:                    "/deposit",
			input_request:               model.Transfer_request{Amount: 50, Category: " 	WOrk	"},
			token:                       user1_token,
			sender_user:                 user1,
			receiver_user:               model.User{},
			expected_status:             200,
			expected_transaction:        model.Transaction{WalletID: user1_wallet.ID, Amount: 50, Type: "deposit", Category: "work", RelatedWalletID: nil},
			expected_sender_balance:     100,
			expected_sender_transaction: model.Transaction{WalletID: user1_wallet.ID, Amount: 50, Type: "deposit", Category: "work", RelatedWalletID: nil},
		}, {
			name:                        "Transfer - Receiver not found",
			endpoint:                    "/transfer",
			input_request:               model.Transfer_request{Amount: 50, Category: " 	food	", ToUsername: "	Hasan	"},
			token:                       user1_token,
			sender_user:                 user1,
			receiver_user:               model.User{},
			expected_status:             404,
			expected_transaction:        model.Transaction{},
			expected_sender_balance:     100,
			expected_sender_transaction: model.Transaction{},
		}, {
			name:                        "Transfer - Error Self Transfer",
			endpoint:                    "/transfer",
			input_request:               model.Transfer_request{Amount: 50, Category: " 	food	", ToUsername: "	ali	"},
			token:                       user1_token,
			sender_user:                 user1,
			receiver_user:               user1,
			expected_status:             400,
			expected_transaction:        model.Transaction{},
			expected_sender_balance:     100,
			expected_sender_transaction: model.Transaction{},
		}, {
			name:                        "deposit - Invalid amount",
			endpoint:                    "/deposit",
			input_request:               model.Transfer_request{Amount: -20, Category: " 	WOrk	"},
			token:                       user1_token,
			sender_user:                 user1,
			receiver_user:               model.User{},
			expected_status:             400,
			expected_transaction:        model.Transaction{},
			expected_sender_balance:     100,
			expected_sender_transaction: model.Transaction{},
		}, {
			name:                        "withdraw - Invalid token",
			endpoint:                    "/withdraw",
			input_request:               model.Transfer_request{Amount: 100, Category: " 	FoOd	"},
			token:                       "",
			sender_user:                 user1,
			receiver_user:               model.User{},
			expected_status:             401,
			expected_transaction:        model.Transaction{},
			expected_sender_balance:     100,
			expected_sender_transaction: model.Transaction{},
		}, {
			name:                        "Transfer - Insufficient fund",
			endpoint:                    "/transfer",
			input_request:               model.Transfer_request{Amount: 150, Category: " 	FoOd	", ToUsername: "	mOHamed	"},
			token:                       user1_token,
			sender_user:                 user1,
			receiver_user:               user2,
			expected_status:             400,
			expected_transaction:        model.Transaction{},
			expected_sender_balance:     100,
			expected_sender_transaction: model.Transaction{},
		},
	}
	//run tests one by one.
	for _, usecase := range tests {
		t.Run(usecase.name, func(t *testing.T) {
			//set the json body.
			json_request, err := json.Marshal(usecase.input_request)
			if err != nil {
				t.Fatalf("Filed to convert request to json format in case: %s", usecase.name)
			}
			//set the request.
			request, err := http.NewRequest("POST", "/wallet"+usecase.endpoint, bytes.NewBuffer(json_request))
			if err != nil {
				t.Fatalf("Failed to set the request using json body in case: %s", usecase.name)
			}
			//set the request header.
			request.Header.Set("Content-Type", "application/json")
			//set the token.
			if usecase.token != "" {
				request.Header.Set("Authorization", "Bearer "+usecase.token)
			}
			//set a receiver to response.
			receiver := httptest.NewRecorder()
			//pass request to the router.
			Router.ServeHTTP(receiver, request)
			//check the response.
			if usecase.expected_status != receiver.Code {
				t.Errorf("Response(case %s): expected status code %d got %d", usecase.name, usecase.expected_status, receiver.Code)
			}
			//check the user wallet after process.
			sender_wallet := Get_wallet_by_username(t, usecase.sender_user)
			if sender_wallet.Balance != usecase.expected_sender_balance {
				t.Errorf("Logic (case %s): expected sender balance after transaction %d got %d", usecase.name, usecase.expected_sender_balance, sender_wallet.Balance)
			}
			//in case of success check the transactions in response and database.
			if usecase.expected_status == 200 && receiver.Code == usecase.expected_status {
				//get the transaction in response.
				var received_transaction model.Transaction
				err := json.Unmarshal(receiver.Body.Bytes(), &received_transaction)
				if err != nil {
					t.Fatalf("Failed to extract transaction from response in case %s", usecase.name)
				}
				//compare expected transaction with received transaction.
				Compare_transactions(t, "Response - "+usecase.name, received_transaction, usecase.expected_transaction)
				//check the stored transaction with the expected one.
				sender_transaction := Get_latest_transaction(t, sender_wallet)
				Compare_transactions(t, "stored sender transaction - "+usecase.name, sender_transaction, usecase.expected_sender_transaction)
				//check if the process was successful transfer.
				if usecase.endpoint == "/transfer" {
					//check the receiver balance.
					receiver_wallet := Get_wallet_by_username(t, usecase.receiver_user)
					if receiver_wallet.Balance != usecase.expected_receiver_balance {
						t.Errorf("Logic (case %s): expected receiver balance after transaction %d got %d", usecase.name, usecase.expected_receiver_balance, receiver_wallet.Balance)
					}
					//check the receiver transaction.
					receiver_transaction := Get_latest_transaction(t, receiver_wallet)
					Compare_transactions(t, "stored receiver transaction - "+usecase.name, receiver_transaction, usecase.expected_receiver_transaction)
				}
			}
		})
	}
}

func Compare_transactions(t *testing.T, stage string, received, expected model.Transaction) {
	//compare the walletID.
	if received.WalletID != expected.WalletID {
		t.Errorf("Logic (%s): expected walletIdD in transaction %d got %d", stage, expected.WalletID, received.WalletID)
	}
	//compare the type of transaction.
	if received.Type != expected.Type {
		t.Errorf("Logic (%s): expected Type in transaction %s got %s", stage, expected.Type, received.Type)
	}
	//compare the amount of transaction.
	if received.Amount != expected.Amount {
		t.Errorf("Logic (%s): expected amount in transaction %d got %d", stage, expected.Amount, received.Amount)
	}
	//compare the category of transaction.
	if received.Category != expected.Category {
		t.Errorf("Logic (%s): expected category in transaction %s got %s", stage, expected.Category, received.Category)
	}
	//check the related wallet ID.
	if received.RelatedWalletID == nil || expected.RelatedWalletID == nil {
		if received.RelatedWalletID == nil && expected.RelatedWalletID != nil {
			t.Errorf("Boundary (%s): expected not nil related wallet id got %v", stage, received.RelatedWalletID)
		} else if received.RelatedWalletID != nil && expected.RelatedWalletID == nil {
			t.Errorf("Boundary (%s): expected nil related wallet ID got %v", stage, received.RelatedWalletID)
		}
	} else {
		//compare the id in expected transaction with actual one.
		if *received.RelatedWalletID != *expected.RelatedWalletID {
			t.Errorf("Logic (%s): expected related wallet ID in transaction %d got %d", stage, *expected.RelatedWalletID, *received.RelatedWalletID)
		}
	}
}

func Test_concurrent_withdrawals(t *testing.T) {

	//first: reset wallets balance to assure balance after operations.
	Reset_wallets_balance(t, 100)
	//then get a token for user to pass authentication middleware.
	user2_token := Login_test_user(t, user2_Auth)
	user2_wallet := Get_wallet_by_username(t, user2)
	//set a channel to get the two responses from concurrent requests.
	var group sync.WaitGroup
	//create a struct for response.
	type response struct {
		code int
		body string
	}
	response_channel := make(chan response, 2)

	//set the request.
	request_data := model.Transfer_request{Amount: 100, Category: "	 	PurCHaSe	"}
	expected_transaction := model.Transaction{WalletID: user2_wallet.ID, Amount: 100, Type: "withdraw", Category: "purchase", RelatedWalletID: nil}
	expected_balance := 0
	//create the json request.
	json_request, err := json.Marshal(request_data)
	if err != nil {
		t.Fatalf("Filed to convert request to json format in concurrent withdrawals")
	}
	//create two concurrent requests to the router.
	for i := 0; i < 2; i++ {
		group.Add(1)
		//create the goroutine.
		go func() {
			defer group.Done()
			//create the request.
			request, err := http.NewRequest("POST", "/wallet/withdraw", bytes.NewBuffer(json_request))
			if err != nil {
				t.Errorf("Failed to set the request using json body in cuncurrent withdrawals")
				return
			}
			//set the request header.
			request.Header.Set("Content-Type", "application/json")
			//set the token.
			request.Header.Set("Authorization", "Bearer "+user2_token)
			//set a receiver to response.
			receiver := httptest.NewRecorder()
			//pass request to the router.
			Router.ServeHTTP(receiver, request)
			//get the response out of goroutine.
			response_channel <- response{
				code: receiver.Code,
				body: receiver.Body.String(),
			}
		}()
	}
	//wait the request to be done.
	group.Wait()
	close(response_channel)
	//check the result of requests.
	var success, fail int
	var successful_transaction model.Transaction
	//chcek the success and fail requests.
	for response := range response_channel {
		if response.code == 200 {
			success++
			//get the transaction in response.
			err := json.Unmarshal([]byte(response.body), &successful_transaction)
			if err != nil {
				t.Fatalf("Failed to extract the transaction from response in concurrent withdrawals")
			}
		} else {
			fail++
		}
	}
	//check the success and fail requests count.
	if success != 1 || fail != 1 {
		t.Errorf("Concurrent withdrawals failed: expected only one success and one failure got %d successes and %d failures", success, fail)
	}
	//compare the transaction in return with expected one.
	Compare_transactions(t, "Response - Concurrent withdrawals", successful_transaction, expected_transaction)
	//check the balance and assure it is zero.
	wallet := Get_wallet_by_username(t, user2)
	if wallet.Balance != expected_balance {
		t.Errorf("Logic (Concurrent withdrawals): expected wallet balance after process %d got %d", expected_balance, wallet.Balance)
	}
	//check the transaction stored in the database.
	transaction := Get_latest_transaction(t, wallet)
	Compare_transactions(t, "Stored transaction - Concurrent withdrawals", transaction, expected_transaction)
}
