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
	Create_user_record(tx *gorm.DB, new_user *model.User) error 
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

func (repo *user_repo) Create_user_record(tx *gorm.DB, new_user *model.User) error {
	//create the user.
	result := tx.Create(new_user)
	//check the result.
	switch {
	case result.Error == nil:
		return nil
	case errors.Is(result.Error, gorm.ErrDuplicatedKey):
		return ErrDuplicatedName
	default:
		return result.Error
	}
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
