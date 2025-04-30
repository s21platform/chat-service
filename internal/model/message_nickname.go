package model

type UpdateNicknameMessage struct {
	UUID     string `json:"uuid"`
	Nickname string `json:"nickname"`
}
