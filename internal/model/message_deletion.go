package model

import "github.com/google/uuid"

type DeletionScope int

const (
	Self DeletionScope = iota
	All
)

type MessageToDelete struct {
	MessageUUID uuid.UUID     `db:"id"` // UUID сообщения
	Scope       DeletionScope //  Область удаления (SELF или ALL)
}

type DeletionResult struct {
	DeletionStatus bool `db:"deleted"` // Статус успешного удаления
}
