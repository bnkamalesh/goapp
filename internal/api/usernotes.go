package api

import (
	"context"
	"errors"

	"github.com/bnkamalesh/goapp/internal/usernotes"
)

func (a *API) CreateUserNote(ctx context.Context, un *usernotes.Note) (*usernotes.Note, error) {
	return nil, errors.New("create user not is not implemented")
}

func (a *API) ReadUserNote(ctx context.Context, userID string, noteID string) (*usernotes.Note, error) {
	return nil, errors.New("read user not is not implemented")
}
