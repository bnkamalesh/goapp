package users

import (
	"context"
	"strings"

	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/goapp/internal/pkg/logger"
)

var (
	ErrUserEmailNotFound      = errors.New("user with the email not found")
	ErrUserEmailAlreadyExists = errors.New("user with the email already exists")
)

type User struct {
	ID             string
	FullName       string
	Email          string
	Phone          string
	ContactAddress string
}

// ValidateForCreate runs the validation required for when a user is being created. i.e. ID is not available
func (us *User) ValidateForCreate() error {
	if us.FullName == "" {
		return errors.Validation("full name cannot be empty")
	}

	if us.Email == "" {
		return errors.Validation("email cannot be empty")
	}

	return nil
}

func (us *User) Sanitize() {
	us.ID = strings.TrimSpace(us.ID)
	us.FullName = strings.TrimSpace(us.FullName)
	us.ContactAddress = strings.TrimSpace(us.ContactAddress)
	us.Phone = strings.TrimSpace(us.Phone)
}

type store interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	SaveUser(ctx context.Context, user *User) (string, error)
	BulkSaveUser(ctx context.Context, users []User) error
}
type Users struct {
	store store
}

func (us *Users) CreateUser(ctx context.Context, user *User) (*User, error) {
	user.Sanitize()
	err := user.ValidateForCreate()
	if err != nil {
		return nil, err
	}

	newID, err := us.store.SaveUser(ctx, user)
	if err != nil {
		return nil, err
	}
	user.ID = newID

	return user, nil
}

func (us *Users) ReadByEmail(ctx context.Context, email string) (*User, error) {
	if email == "" {
		return nil, errors.Validation("no email provided")
	}

	return us.store.GetUserByEmail(ctx, email)
}

func (us *Users) AsyncCreateUsers(ctx context.Context, users []User) error {
	errList := make([]error, 0, len(users))
	for i := range users {
		err := users[i].ValidateForCreate()
		if err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) != 0 {
		return errors.Join(errList...)
	}

	go func() {
		ctx := context.TODO()
		err := us.store.BulkSaveUser(context.TODO(), users)
		if err != nil {
			logger.Error(ctx, err, users)
		}
	}()

	return nil
}

func NewService(store store) *Users {
	return &Users{
		store: store,
	}
}
