package test

import (
	"Wallet-App/model"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func Create_test_user(user model.Auth_request) {
	//define a struct for request body.
	request_body := user
	//set a variable for expected code.
	var expected_code = 201
	//convert struct to json format.
	json_request, _ := json.Marshal(request_body)
	//create the request with the json body.
	req, _ := http.NewRequest("POST", "/signup", bytes.NewBuffer(json_request))
	req.Header.Set("Content-Type", "application/json")
	//set the response receiver.
	receiver := httptest.NewRecorder()
	//move request to the routers.
	Router.ServeHTTP(receiver, req)
	//check the response.
	if receiver.Code != expected_code {
		panic("Failed to setup users for testing")
	}
}

func Login_test_user(t *testing.T, user model.Auth_request) string {
	//define a struct for request body.
	request_body := user
	//set a variable for expected code.
	var expected_code = 200
	//convert struct to json format.
	json_request, _ := json.Marshal(request_body)
	//create the request with the json body.
	req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(json_request))
	req.Header.Set("Content-Type", "application/json")
	//set the response receiver.
	receiver := httptest.NewRecorder()
	//move request to the routers.
	Router.ServeHTTP(receiver, req)
	//check the response.
	if receiver.Code != expected_code {
		t.Fatalf("Failed to setup users for testing: expecetd status code %d got %d", expected_code, receiver.Code)
	}
	//get the token in response.
	var response struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal(receiver.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to parse login token in response")
	}
	//check if the token was invalid.
	if response.Token == "" {
		t.Fatalf("Login token is invalid: expected valid token got empty token")
	}
	return response.Token
}

func Reset_wallets_balance(t *testing.T, initial_balance int) {
	//update all wallets in database.
	result := DB.Exec("UPDATE wallets SET balance = ?", initial_balance)
	if result.Error != nil {
		t.Fatalf("Failed to reset wallets balance to in initial state")
	}
}

func Get_wallet_by_username(t *testing.T, user model.User) model.Wallet {
	//define a variable for target user.
	var target_user model.User
	//get the userID first from database.
	result := DB.Select("id").Where("username = ?", user.Username).First(&target_user)
	if result.Error != nil {
		t.Fatalf("Failed to fetch user from database with its username")
	}
	//define a variable for target wallet.
	var wallet model.Wallet
	result = DB.Where("user_id  = ?", target_user.ID).First(&wallet)
	if result.Error != nil {
		t.Fatalf("Failed to fetch wallet with the user ID")
	}
	//return the wallet.
	return wallet
}

func Get_latest_transaction(t *testing.T, wallet model.Wallet) model.Transaction {
	//define a struct for the transaction.
	var transaction model.Transaction
	//get the latest transaction.
	result := DB.Where("wallet_id = ?", wallet.ID).Order("id desc").First(&transaction)
	//check the result.
	if result.Error != nil {
		t.Fatalf("Failed to fetch the latest transaction for the wallet ID")
	}
	return transaction
}
