package api

import (
	"context"
	"time"

	"github.com/bnkamalesh/goapp/internal/usernotes"
	"github.com/bnkamalesh/goapp/internal/users"
)

var (
	now = time.Now()
)

// Server has all the methods required to run the server
type Server interface {
	CreateUser(ctx context.Context, user *users.User) (*users.User, error)
	ReadUserByEmail(ctx context.Context, email string) (*users.User, error)
	CreateUserNote(ctx context.Context, un *usernotes.Note) (*usernotes.Note, error)
	ReadUserNote(ctx context.Context, userID string, noteID string) (*usernotes.Note, error)
	ServerHealth() (map[string]any, error)
}

// Subscriber has all the methods required to run the subscriber
type Subscriber interface {
	AsyncCreateUsers(ctcx context.Context, users []users.User) error
}

type API struct {
	users  *users.Users
	unotes *usernotes.UserNotes
}

// ServerHealth returns the health of the serever app along with other info like version
func (a *API) ServerHealth() (map[string]any, error) {
	return map[string]any{
		"env":        "testing",
		"version":    "v0.1.0",
		"commit":     "<git commit hash>",
		"status":     "all systems up and running",
		"startedAt":  now.String(),
		"releasedOn": now.String(),
	}, nil

}

func New(us *users.Users, un *usernotes.UserNotes) *API {
	return &API{
		users:  us,
		unotes: un,
	}
}

func NewServer(us *users.Users, un *usernotes.UserNotes) Server {
	return New(us, un)
}

func NewSubscriber(us *users.Users) Subscriber {
	return New(us, nil)
}
