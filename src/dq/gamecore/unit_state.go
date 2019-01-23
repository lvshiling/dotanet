package gamecore

type UnitState interface {
	OnTransform()
	Update(dt float64)
	OnEnd()
	OnStart()
}

type IdleState struct {
	Parent *Unit
}

func NewIdleState(p *Unit) *IdleState {
	re := &IdleState{}
	re.Parent = p
	return re
}

func (this *IdleState) OnTransform() {

}
func (this *IdleState) Update(dt float64) {

}
func (this *IdleState) OnEnd() {

}
func (this *IdleState) OnStart() {

}
