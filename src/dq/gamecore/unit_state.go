package gamecore

import (
	"dq/log"
	"dq/protobuf"
	"dq/vec2d"
)

type UnitState interface {
	OnTransform()
	Update(dt float64)
	OnEnd()
	OnStart()
}

//------------------------------休息状态-------------------------
type IdleState struct {
	Parent *Unit
}

func NewIdleState(p *Unit) *IdleState {
	log.Info(" NewIdleState")
	re := &IdleState{}
	re.Parent = p
	re.OnStart()
	return re
}

//检查状态变换
func (this *IdleState) OnTransform() {
	move := this.Parent.Move
	//切换到移动状态
	if move != nil {
		if move.IsStart == true && this.Parent.GetCanMove() == true {
			this.OnEnd()
			this.Parent.State = NewMoveState(this.Parent)
		}
	}
}
func (this *IdleState) Update(dt float64) {

}
func (this *IdleState) OnEnd() {

}
func (this *IdleState) OnStart() {
	//this.OnTransform()
	this.Parent.AnimotorState = 1
}

//------------------------------移动状态-------------------------
type MoveState struct {
	Parent *Unit

	MoveData *protomsg.CS_PlayerMove
}

func NewMoveState(p *Unit) *MoveState {

	log.Info(" NewMoveState")
	re := &MoveState{}
	re.Parent = p
	re.OnStart()
	return re
}

//检查状态变换
func (this *MoveState) OnTransform() {
	//move := this.Parent.Move
	//切换到idle状态
	//if this.MoveData == nil || this.MoveData.IsStart == false || this.Parent.GetCanMove() == false {
	//	if this.MoveData == nil || this.MoveData.IsStart == false || this.Parent.GetCanMove() == false {
	//		this.OnEnd()
	//		this.Parent.State = NewIdleState(this.Parent)

	//		log.Info(" 111111")
	//	}
	if this.Parent.Body.IsMove() == false {
		this.OnEnd()
		this.Parent.State = NewIdleState(this.Parent)

		log.Info(" 222222")
	}
}
func (this *MoveState) Update(dt float64) {
	move := this.Parent.Move

	if move != nil && this.MoveData != move {
		log.Info(" move:%v----%v---%d---%d", move, this.MoveData, move, this.MoveData)
		this.MoveData = move
		this.Parent.Body.SetTarget(vec2d.Vec2{X: float64(this.MoveData.X), Y: float64(this.MoveData.Y)})
	}

}
func (this *MoveState) OnEnd() {

}
func (this *MoveState) OnStart() {
	//this.OnTransform()
	this.MoveData = this.Parent.Move
	if this.MoveData.IsStart == false {
		return
	}
	this.Parent.Move = nil
	this.Parent.AnimotorState = 2
	this.Parent.Body.SetTarget(vec2d.Vec2{X: float64(this.MoveData.X), Y: float64(this.MoveData.Y)})
}
