package http

import (
	"encoding/json"
	"net/http"

	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/webgo/v7"

	"github.com/bnkamalesh/goapp/internal/users"
)

// CreateUser is the HTTP handler to create a new user
// This handler does not use any framework, instead just the standard library
func (h *Handlers) CreateUser(w http.ResponseWriter, r *http.Request) error {
	u := new(users.User)
	err := json.NewDecoder(r.Body).Decode(u)
	if err != nil {
		return errors.InputBodyErr(err, "invalid JSON provided")
	}

	createdUser, err := h.apis.CreateUser(r.Context(), u)
	if err != nil {
		return err
	}

	webgo.R200(w, createdUser)

	return nil
}

// ReadUserByEmail is the HTTP handler to read an existing user by email
func (h *Handlers) ReadUserByEmail(w http.ResponseWriter, r *http.Request) error {
	wctx := webgo.Context(r)
	email := wctx.Params()["email"]

	out, err := h.apis.ReadUserByEmail(r.Context(), email)
	if err != nil {
		return err
	}

	webgo.R200(w, out)

	return nil
}
