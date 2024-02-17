package usernotes

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/bnkamalesh/errors"
	"github.com/bnkamalesh/goapp/internal/users"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type pgstore struct {
	qbuilder  squirrel.StatementBuilderType
	pqdriver  *pgxpool.Pool
	tableName string
}

func (ps *pgstore) GetNoteByID(ctx context.Context, userID string, noteID string) (*Note, error) {
	query, args, err := ps.qbuilder.Select(
		"title",
		"content",
		"created_at",
		"updated_at",
	).From(
		ps.tableName,
	).Where(
		squirrel.Eq{
			"id":      noteID,
			"user_id": userID,
		},
	).ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "failed preparing query")
	}

	usernote := &Note{
		ID: noteID,
		Creator: &users.User{
			ID: userID,
		},
	}

	err = ps.pqdriver.QueryRow(
		ctx, query, args...,
	).Scan(
		&usernote.Title,
		&usernote.Content,
		&usernote.CreatedAt,
		&usernote.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "failed getting user note")
	}

	return usernote, nil
}

func (ps *pgstore) SaveNote(ctx context.Context, note *Note) (string, error) {
	noteID := ps.newNoteID()
	query, args, err := ps.qbuilder.Insert(
		ps.tableName,
	).Columns(
		"id",
		"title",
		"content",
		"user_id",
	).Values(
		note.ID,
		note.Title,
		note.Content,
		note.Creator.ID,
	).ToSql()
	if err != nil {
		return "", errors.Wrap(err, "failed preparing query")
	}

	_, err = ps.pqdriver.Exec(ctx, query, args...)
	if err != nil {
		return "", errors.Wrap(err, "failed storing note")
	}

	return noteID, nil
}

func (ps *pgstore) newNoteID() string {
	return uuid.New().String()
}

func NewPostgresStore(pqdriver *pgxpool.Pool, tableName string) store {
	return &pgstore{
		qbuilder:  squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
		pqdriver:  pqdriver,
		tableName: tableName,
	}
}
