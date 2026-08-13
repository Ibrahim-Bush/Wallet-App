package repository

import (
	"Wallet-App/model"
	"errors"

	"gorm.io/gorm"
)

var (
	ErrDuplicatedName = errors.New("Username already exists")
	ErrUserNotFound   = errors.New("User not found")
)

type User_repo interface {
	Create_user_record(new_user *model.User, wallet *model.Wallet) error
	Get_user_record_by_name(name string) (*model.User, error)
	Get_user_record_by_id(id int) (*model.User, error)
}

type user_repo struct {
	db *gorm.DB
}

func Init_user_repo(db *gorm.DB) User_repo {
	new_repo := user_repo{db: db}
	return &new_repo
}

func (repo *user_repo) Create_user_record(new_user *model.User, wallet *model.Wallet) error {
	//Create a user with wallet.
	err := repo.db.Transaction(func(tx *gorm.DB) error {
		//first: create the user.
		result := tx.Create(new_user)
		//check the result.
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
				return ErrDuplicatedName
			}
			return result.Error
		}
		//if the user created successfully, create his wallet.
		wallet.UserID = new_user.ID
		result = tx.Create(wallet)
		//check the result.
		if result.Error != nil {
			if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
				return ErrDuplicatedWallet
			}
			return result.Error
		}
		//if both user and wallet creation succeeded, return nil.
		return nil
	})
	//return the result.
	return err
}

func (repo *user_repo) Get_user_record_by_name(name string) (*model.User, error) {
	//search for username in database.
	var user model.User
	result := repo.db.Where("username = ?", name).First(&user)
	//check the result of process.
	switch {
	case result.Error == nil:
		return &user, nil
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return nil, ErrUserNotFound
	default:
		return nil, result.Error
	}
}

func (repo *user_repo) Get_user_record_by_id(id int) (*model.User, error) {
	//search for id in database.
	var user model.User
	result := repo.db.First(&user, id)
	//check the result of process.
	switch {
	case result.Error == nil:
		return &user, nil
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return nil, ErrUserNotFound
	default:
		return nil, result.Error
	}
}
