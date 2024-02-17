package goapp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/bnkamalesh/errors"
)

type User struct {
	ID             string
	FullName       string
	Email          string
	Phone          string
	ContactAddress string
}

type GoApp struct {
	client    *http.Client
	basePath  string
	usersBase string
}

func (ht *GoApp) makeRequest(ctx context.Context, req *http.Request) ([]byte, error) {
	resp, err := ht.client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed making request")
	}

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed reading response body")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("%d: %s", resp.StatusCode, string(raw))
	}

	return raw, nil
}

func (ht *GoApp) CreateUser(ctx context.Context, us *User) (*User, error) {
	payload, err := json.Marshal(us)
	if err != nil {
		return nil, errors.Wrap(err, "failed marshaling to json")
	}
	buff := bytes.NewBuffer(payload)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		ht.usersBase,
		buff,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed preparing request")
	}

	raw, err := ht.makeRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	respUsr := struct {
		Data User `json:"data"`
	}{}
	err = json.Unmarshal(raw, &respUsr)
	if err != nil {
		return nil, errors.Wrap(err, "failed unmarshaling user")
	}

	return &respUsr.Data, nil
}

func (ht *GoApp) UserByEmail(ctx context.Context, email string) (*User, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/%s", ht.usersBase, email),
		nil,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed preparing request")
	}

	raw, err := ht.makeRequest(ctx, req)
	if err != nil {
		return nil, err
	}

	respUsr := struct {
		Data User `json:"data"`
	}{}
	err = json.Unmarshal(raw, &respUsr)
	if err != nil {
		return nil, errors.Wrap(err, "failed unmarshaling user")
	}

	return &respUsr.Data, nil
}

func NewClient(basePath string) *GoApp {
	return &GoApp{
		client:    http.DefaultClient,
		basePath:  basePath,
		usersBase: basePath + "/users",
	}
}
