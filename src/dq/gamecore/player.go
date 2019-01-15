package gamecore

type Player struct {
	Uid       int32
	ConnectId int32
	MainUnit  *Unit //主单位

	//OtherUnit  *Unit //其他单位
}

func CreatePlayer(uid int32, connectid int32) *Player {
	re := &Player{}
	re.Uid = uid
	re.ConnectId = connectid

	return re
}
