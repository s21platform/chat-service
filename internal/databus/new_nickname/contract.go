package new_nickname

type DBRepo interface {
	UpdateUserNickname(userUUID, newNickname string) error
}
