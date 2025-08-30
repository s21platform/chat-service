package model

type StreamMember struct {
	UserID   string
	Metadata string
}

type StreamMemberParams struct {
	UserID    string `db:"id"`
	Nickname  string `db:"nickname"`
	AvatarURL string `db:"avatar_url"`
}

type UserSubscription struct {
	UserID  string
	Channel string
}
