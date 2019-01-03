package network

type Agent interface {
	Run()
	OnClose()
	GetConnectId() int32
	GetModeType() string
}
