package mock

import "context"

const (
	TestUserNormal = "standard-user"
)

type UserModel struct{}

func (model *UserModel) Exists(_ context.Context, id string) (bool, error) {
	switch id {
	case TestUserNormal:
		return true, nil
	default:
		return false, nil
	}
}

func (model *UserModel) RecordLogIn(context.Context, string) (bool, error) {
	return false, nil
}

func (model *UserModel) UpdateName(context.Context, string, string) error {
	return nil
}
