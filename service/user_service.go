package service

import (
	"Wallet-App/model"
	"Wallet-App/repository"
	"Wallet-App/utils"
	"errors"
	"strings"
)

var (
	ErrEmptyName        = errors.New("Username cannot be empty")
	ErrEmptyPassword    = errors.New("Password cannot be empty")
	ErrInvalidPassword  = errors.New("Invalid password")
	ErrDuplicatedName   = errors.New("Username already exists")
	ErrDuplicatedWallet = errors.New("This user has a wallet already")
	ErrUserNotFound     = errors.New("User not found")
	ErrWrongPassword    = errors.New("Incorrect password")
	ErrServerError      = errors.New("Server error")
)

type User_service interface {
	Create_user(input_struct model.Auth_request) (*model.User, *model.Wallet, error)
	Login_user(input_struct model.Auth_request) (string, error)
}

type user_service struct {
	repository  repository.User_repo
	wallet_repo repository.Wallet_repo
}

func Init_user_service(repo repository.User_repo, wallet_repo repository.Wallet_repo) User_service {
	new_service := user_service{repository: repo, wallet_repo: wallet_repo}
	return &new_service
}

func (service *user_service) Create_user(input_struct model.Auth_request) (*model.User, *model.Wallet, error) {
	//normalize input fields.
	input_struct.Username = strings.ToLower(strings.TrimSpace(input_struct.Username))
	input_struct.Password = strings.TrimSpace(input_struct.Password)
	//check if there is empty field.
	if input_struct.Username == "" {
		return nil, nil, ErrEmptyName
	}
	if input_struct.Password == "" {
		return nil, nil, ErrEmptyPassword
	}
	//hash password before storing it.
	hashed_Password, err := utils.Hash_password(input_struct.Password)
	//check the result.
	if err != nil {
		return nil, nil, ErrInvalidPassword
	}
	//create a new object for user.
	var user = model.User{
		Username: input_struct.Username,
		Password: hashed_Password,
		Role:     "user",
	}
	//create a wallet for the user.
	wallet := model.Wallet{
		Balance: 0,
	}
	//store new user in repository.
	err = service.repository.Create_user_record(&user, &wallet)
	//check the result.
	switch {
	case err == nil:
		return &user, &wallet, nil
	case errors.Is(err, repository.ErrDuplicatedName):
		return nil, nil, ErrDuplicatedName
	case errors.Is(err, repository.ErrDuplicatedWallet):
		return nil, nil, ErrDuplicatedWallet
	default:
		return nil, nil, ErrServerError
	}
}

func (service *user_service) Login_user(input_struct model.Auth_request) (string, error) {
	//first normalize username.
	username := strings.ToLower(strings.TrimSpace(input_struct.Username))
	password := strings.TrimSpace(input_struct.Password)
	//search for user in repository.
	user, err := service.repository.Get_user_record_by_name(username)
	//check the result.
	if err != nil {
		//check if it was server error or not found.
		if errors.Is(err, repository.ErrUserNotFound) {
			return "", ErrUserNotFound
		}
		return "", ErrServerError
	}
	//if the user exists compare the password.
	result := utils.Compare_password(password, user.Password)
	//if the password is wrong.
	if !result {
		return "", ErrWrongPassword
	}
	//if password is true, create a token.
	claims := model.User_claims{
		Username: user.Username,
		UserID:   user.ID,
		Role:     user.Role,
	}
	token, err := utils.Generate_token(claims)
	if err != nil {
		return "", ErrServerError
	}
	return token, nil
}
