package model

type ChatMemberParams struct {
	UserUUID   string `db:"user_uuid"`
	Nickname   string `db:"nickname"`
	AvatarLink string `db:"avatar_link"`
}
