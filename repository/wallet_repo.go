package repository

import (
	"Wallet-App/model"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	ErrDuplicatedWallet = errors.New("Wallet already exists")
	ErrWalletNotFound   = errors.New("Wallet not found")
)

type Wallet_repo interface {
	Create_wallet_record(tx *gorm.DB, new_wallet *model.Wallet) error
	Get_record_by_userID(id int) (*model.Wallet, error)
	Get_all_records() ([]model.Wallet, error)
	Get_record_by_userID_with_lock(tx *gorm.DB, id int) (*model.Wallet, error)
	Update_wallet_balance(tx *gorm.DB, wallet_id, new_balance int) error
}

func Init_wallet_repo(db *gorm.DB) Wallet_repo {
	new_repo := wallet_repo{db: db}
	return &new_repo
}

type wallet_repo struct {
	db *gorm.DB
}

func (wallet *wallet_repo) Create_wallet_record(tx *gorm.DB, new_wallet *model.Wallet) error {
	//check the transaction object.
	if tx == nil {
		tx = wallet.db
	}
	//store the wallet in the database.
	result := tx.Create(new_wallet)
	//check the result.
	switch {
	case result.Error == nil:
		return nil
	case errors.Is(result.Error, gorm.ErrDuplicatedKey):
		return ErrDuplicatedWallet
	default:
		return result.Error
	}
}

func (wallet *wallet_repo) Get_record_by_userID(id int) (*model.Wallet, error) {
	//create a wallet variable for the target.
	var target model.Wallet
	//search for desired wallet.
	result := wallet.db.Where("user_id = ?", id).First(&target)
	//check the result.
	switch {
	case result.Error == nil:
		return &target, nil
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return nil, ErrWalletNotFound
	default:
		return nil, result.Error
	}
}

func (wallet *wallet_repo) Get_all_records() ([]model.Wallet, error) {
	//create a slice for records.
	var wallets = make([]model.Wallet, 0)
	//get all wallets in the system.
	result := wallet.db.Find(&wallets)
	switch {
	case result.Error == nil:
		return wallets, nil
	default:
		return nil, result.Error
	}
}

func (wallet *wallet_repo) Get_record_by_userID_with_lock(tx *gorm.DB, id int) (*model.Wallet, error) {
	//create a wallet variable for the target.
	var target model.Wallet
	//get the target wallet.
	result := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", id).First(&target)
	//check the result.
	switch {
	case result.Error == nil:
		return &target, nil
	case errors.Is(result.Error, gorm.ErrRecordNotFound):
		return nil, ErrWalletNotFound
	default:
		return nil, result.Error
	}
}

func (wallet *wallet_repo) Update_wallet_balance(tx *gorm.DB, wallet_id, new_balance int) error {
	//update the balance of the target wallet.
	result := tx.Model(&model.Wallet{}).Where("id = ?", wallet_id).Update("balance", new_balance)
	//check the result of process.
	if result.Error != nil {
		return result.Error
	} else if result.RowsAffected == 0 {
		return ErrWalletNotFound
	} else {
		return nil
	}
}
