package users

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/bnkamalesh/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userStore struct {
	qbuilder  squirrel.StatementBuilderType
	pqdriver  *pgxpool.Pool
	tableName string
}

func (us *userStore) Create(ctx context.Context, u *User) error {
	query, args, err := us.qbuilder.Insert(us.tableName).SetMap(map[string]interface{}{
		"firstName": u.FirstName,
		"lastName":  u.LastName,
		"mobile":    u.Mobile,
		"email":     u.Email,
		"createdAt": u.CreatedAt,
		"updatedAt": u.UpdatedAt,
	}).ToSql()
	if err != nil {
		return errors.InternalErr(err, errors.DefaultMessage)
	}

	_, err = us.pqdriver.Exec(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "violates unique constraint") {
			return errors.DuplicateErrf(err, "user with email '%s' already exists", u.Email)
		}
		return errors.InternalErr(err, errors.DefaultMessage)
	}

	return nil
}

func (us *userStore) ReadByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := us.qbuilder.Select(
		"firstName",
		"lastName",
		"mobile",
		"email",
		"createdAt",
		"updatedAt",
	).From(
		us.tableName,
	).Where(
		squirrel.Eq{
			"email": email,
		},
	).ToSql()
	if err != nil {
		return nil, errors.InternalErr(err, errors.DefaultMessage)
	}

	user := new(User)
	firstName := new(sql.NullString)
	lastName := new(sql.NullString)
	mobile := new(sql.NullString)
	storeEmail := new(sql.NullString)

	row := us.pqdriver.QueryRow(ctx, query, args...)
	err = row.Scan(
		firstName,
		lastName,
		mobile,
		storeEmail,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFoundErrf(err, "user with email '%s' not found", email)
		}

		return nil, errors.InternalErr(err, errors.DefaultMessage)
	}

	user.FirstName = firstName.String
	user.LastName = lastName.String
	user.Mobile = mobile.String
	user.Email = storeEmail.String

	return user, nil
}

func NewStore(pqdriver *pgxpool.Pool) (*userStore, error) {
	return &userStore{
		pqdriver:  pqdriver,
		qbuilder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		tableName: "Users",
	}, nil
}
