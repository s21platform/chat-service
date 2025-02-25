package model

type UserInfo struct {
	UserName   string `db:"username"`
	AvatarLink string `db:"avatar_link"`
}
