package api

import (
	"context"

	"github.com/bnkamalesh/goapp/internal/users"
)

// CreateUser is the API to create/signup a new user
func (a *API) CreateUser(ctx context.Context, u *users.User) (*users.User, error) {
	u, err := a.users.CreateUser(ctx, u)
	if err != nil {
		return nil, err
	}

	return u, nil
}

// ReadUserByEmail is the API to read an existing user by their email
func (a *API) ReadUserByEmail(ctx context.Context, email string) (*users.User, error) {
	u, err := a.users.ReadByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (a *API) AsyncCreateUsers(ctx context.Context, users []users.User) error {
	return a.users.AsyncCreateUsers(ctx, users)
}
