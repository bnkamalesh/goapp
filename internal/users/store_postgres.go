package users

import (
	"context"
	"database/sql"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/bnkamalesh/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgstore struct {
	qbuilder  squirrel.StatementBuilderType
	pqdriver  *pgxpool.Pool
	tableName string
}

func (ps *pgstore) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	query, args, err := ps.qbuilder.Select(
		"id",
		"full_name",
		"email",
		"phone",
		"contact_address",
	).From(
		ps.tableName,
	).Where(
		squirrel.Eq{"email": email},
	).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed preparing query")
	}

	user := new(User)
	uid := new(uuid.NullUUID)
	address := new(sql.NullString)
	phone := new(sql.NullString)

	row := ps.pqdriver.QueryRow(ctx, query, args...)
	err = row.Scan(uid, &user.FullName, &user.Email, phone, address)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.NotFoundErr(ErrUserEmailNotFound, email)
		}
		return nil, errors.Wrap(err, "failed getting user info")
	}
	user.ID = uid.UUID.String()
	user.ContactAddress = address.String
	user.Phone = phone.String

	return user, nil
}

func (ps *pgstore) SaveUser(ctx context.Context, user *User) (string, error) {
	user.ID = ps.newUserID()

	query, args, err := ps.qbuilder.Insert(
		ps.tableName,
	).Columns(
		"id",
		"full_name",
		"email",
		"phone",
		"contact_address",
	).Values(
		user.ID,
		user.FullName,
		user.Email,
		sql.NullString{
			String: user.Phone,
			Valid:  len(user.Phone) != 0,
		},
		sql.NullString{
			String: user.ContactAddress,
			Valid:  len(user.ContactAddress) != 0,
		},
	).ToSql()
	if err != nil {
		return "", errors.Wrap(err, "failed preparing query")
	}
	_, err = ps.pqdriver.Exec(ctx, query, args...)
	if err != nil {
		if strings.Contains(err.Error(), "violates unique constraint \"users_email_key\"") {
			return "", errors.DuplicateErr(ErrUserEmailAlreadyExists, user.Email)
		}
		return "", errors.Wrap(err, "failed storing user info")
	}

	return user.ID, nil
}

func (ps *pgstore) BulkSaveUser(ctx context.Context, users []User) error {
	rows := make([][]any, 0, len(users))

	for _, user := range users {
		rows = append(rows, []any{
			user.ID,
			user.FullName,
			user.Email,
			sql.NullString{
				String: user.Phone,
				Valid:  len(user.Phone) != 0,
			},
			sql.NullString{
				String: user.ContactAddress,
				Valid:  len(user.ContactAddress) != 0,
			},
		})
	}

	inserted, err := ps.pqdriver.CopyFrom(
		ctx,
		pgx.Identifier{ps.tableName},
		[]string{"id", "full_name", "email", "phone", "contact_address"},
		pgx.CopyFromRows(rows),
	)
	if err != nil {
		return errors.Wrap(err, "failed inserting users")
	}

	ulen := int64(len(users))
	if inserted != ulen {
		return errors.Internalf(
			"failed inserting %d out of %d users",
			ulen-inserted,
			ulen,
		)
	}

	return nil
}

func (ps *pgstore) newUserID() string {
	return uuid.NewString()
}

func NewPostgresStore(pqdriver *pgxpool.Pool, tablename string) *pgstore {
	return &pgstore{
		qbuilder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		pqdriver:  pqdriver,
		tableName: tablename,
	}
}
