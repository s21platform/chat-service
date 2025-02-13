package model

import "github.com/google/uuid"

const (
	Self string = "self"
	All  string = "all"
)

type MessageToDelete struct {
	MessageID uuid.UUID `db:"id"` // UUID сообщения
	Mode      string    //  Область удаления (Self или All)
}

type DeletionResult struct {
	DeletionStatus bool `db:"deleted"` // Статус успешного удаления
}
