package mock

import (
	"Wallet-App/model"

	"gorm.io/gorm"
)

type Mock_user_repo struct {
	Received_name string
	Wanted_user   *model.User
	Wanted_err    error
	Was_called    bool
}

func (mock *Mock_user_repo) Get_user_record_by_name(name string) (*model.User, error) {
	//change is called to true.
	mock.Was_called = true
	//store the input data.
	mock.Received_name = name
	//return required value.
	return mock.Wanted_user, mock.Wanted_err
}

func (mock *Mock_user_repo) Create_user_record(tx *gorm.DB, new_user *model.User) error {
	panic("Mock error: unexpected call to Create_user method")
}

func (mock *Mock_user_repo) Get_user_record_by_id(id int) (*model.User, error) {
	panic("Mock error: unexpected call to Get_user_by_id method")
}
