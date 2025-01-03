// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0
// source: person.sql

package repository

import (
	"context"
)

const findPersonByName = `-- name: FindPersonByName :one
SELECT id, full_name FROM person WHERE full_name = $1
`

func (q *Queries) FindPersonByName(ctx context.Context, fullName string) (Person, error) {
	row := q.db.QueryRow(ctx, findPersonByName, fullName)
	var i Person
	err := row.Scan(&i.ID, &i.FullName)
	return i, err
}

const insertPerson = `-- name: InsertPerson :one

INSERT INTO person (id, full_name)
VALUES (uuid_generate_v4(), $1)
RETURNING id, full_name
`

func (q *Queries) InsertPerson(ctx context.Context, fullName string) (Person, error) {
	row := q.db.QueryRow(ctx, insertPerson, fullName)
	var i Person
	err := row.Scan(&i.ID, &i.FullName)
	return i, err
}
