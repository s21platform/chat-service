package model

const (
	Self string = "self"
	All  string = "all"
)

type DeletionInfo struct {
	DeleteFormat string `db:"delete_format"`
	DeletedBy    string `db:"deleted_by"`
	DeletedAt    string `db:"deleted_at"`
}
