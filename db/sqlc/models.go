// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Post struct {
	ID          int64              `json:"id"`
	Filename    string             `json:"filename"`
	DeletionKey string             `json:"deletion_key"`
	Hash        string             `json:"hash"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}
