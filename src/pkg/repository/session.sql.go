// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: session.sql

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const findSessionByID = `-- name: FindSessionByID :one
SELECT id, institution_id, occasion_id, date_time FROM session WHERE id = $1
`

func (q *Queries) FindSessionByID(ctx context.Context, id uuid.UUID) (Session, error) {
	row := q.db.QueryRow(ctx, findSessionByID, id)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.InstitutionID,
		&i.OccasionID,
		&i.DateTime,
	)
	return i, err
}

const insertSession = `-- name: InsertSession :one

INSERT INTO session (id, institution_id, occasion_id, date_time)
VALUES (uuid_generate_v4(), $1, $2, $3)
RETURNING id, institution_id, occasion_id, date_time
`

type InsertSessionParams struct {
	InstitutionID uuid.UUID `json:"institution_id"`
	OccasionID    uuid.UUID `json:"occasion_id"`
	DateTime      time.Time `json:"date_time"`
}

func (q *Queries) InsertSession(ctx context.Context, arg InsertSessionParams) (Session, error) {
	row := q.db.QueryRow(ctx, insertSession, arg.InstitutionID, arg.OccasionID, arg.DateTime)
	var i Session
	err := row.Scan(
		&i.ID,
		&i.InstitutionID,
		&i.OccasionID,
		&i.DateTime,
	)
	return i, err
}
