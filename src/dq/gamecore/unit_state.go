package gamecore

import (
	//"dq/log"
	//"dq/protobuf"
	"dq/utils"
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
	//log.Info(" NewIdleState")
	re := &IdleState{}
	re.Parent = p
	re.OnStart()
	return re
}

//检查状态变换
func (this *IdleState) OnTransform() {

	//攻击命令
	if this.Parent.HaveAttackCmd() {
		//在攻击范围内
		if this.Parent.IsInAttackRange(this.Parent.AttackCmdDataTarget) {
			if this.Parent.AttackCmdDataTarget.IsCanBeAttack() {

				this.OnEnd()
				this.Parent.SetState(NewAttackState(this.Parent))
				return
			} else {

			}
		} else {
			//有攻击指令 却不在攻击范围内
			this.OnEnd()
			this.Parent.SetState(NewMoveState(this.Parent))
			return
		}
	}

	if this.Parent.HaveMoveCmd() && this.Parent.GetCanMove() {
		this.OnEnd()
		this.Parent.SetState(NewMoveState(this.Parent))
		return
	}

}
func (this *IdleState) Update(dt float64) {
	this.Parent.SetAnimotorState(1)
}
func (this *IdleState) OnEnd() {

}
func (this *IdleState) OnStart() {

}

//------------------------------移动状态-------------------------
type MoveState struct {
	Parent *Unit

	LastFindPathTarget     *Unit
	LastFindPathTargetTime float64
}

func NewMoveState(p *Unit) *MoveState {

	//log.Info(" NewMoveState")
	re := &MoveState{}
	re.Parent = p
	re.OnStart()
	return re
}

//检查状态变换
func (this *MoveState) OnTransform() {

	//攻击命令
	if this.Parent.HaveAttackCmd() {
		//在攻击范围内
		if this.Parent.IsInAttackRange(this.Parent.AttackCmdDataTarget) {
			//在范围内 能被攻击就到攻击状态  不能被攻击就到休息状态
			if this.Parent.AttackCmdDataTarget.IsCanBeAttack() {

				this.OnEnd()
				this.Parent.SetState(NewAttackState(this.Parent))
				return
			} else {
				this.OnEnd()
				this.Parent.SetState(NewIdleState(this.Parent))
				return
			}
			return
		}
	} else {
		if this.Parent.HaveMoveCmd() == false || this.Parent.GetCanMove() == false {
			this.OnEnd()
			this.Parent.SetState(NewIdleState(this.Parent))
			return
		}
	}

}

func (this *MoveState) Update(dt float64) {

	//如果速度小于等于0就休息(可能是 寻路失败)
	if this.Parent.Body.CurSpeedSize <= 0 {
		this.Parent.SetAnimotorState(1)
	} else {
		this.Parent.SetAnimotorState(2)
	}

	//先检查攻击对象
	if this.Parent.HaveAttackCmd() {
		//上次寻路的目标单位和本次相同则在1S内 不再寻路
		if this.LastFindPathTarget == this.Parent.AttackCmdDataTarget {
			if utils.GetCurTimeOfSecond()-this.LastFindPathTargetTime < 1 {
				return
			}
		}
		//
		if this.Parent.AttackCmdDataTarget.Body != nil {
			this.Parent.Body.SetTarget(this.Parent.AttackCmdDataTarget.Body.Position)
			this.LastFindPathTarget = this.Parent.AttackCmdDataTarget
			this.LastFindPathTargetTime = utils.GetCurTimeOfSecond()
		}
		return
	}

	//再检查移动命令
	if this.Parent.HaveMoveCmd() {
		this.Parent.Body.SetMoveDir(vec2d.Vec2{X: float64(this.Parent.MoveCmdData.X), Y: float64(this.Parent.MoveCmdData.Y)})
	}

}
func (this *MoveState) OnEnd() {
	this.Parent.Body.ClearMoveDirAndMoveTarget()
}
func (this *MoveState) OnStart() {

	this.LastFindPathTarget = nil
	this.LastFindPathTargetTime = 0

	//this.Parent.Body.SetMoveDir(vec2d.Vec2{X: float64(this.Parent.MoveCmdData.X), Y: float64(this.Parent.MoveCmdData.Y)})
}

//------------------------------攻击状态--------------( 或者攻击)-----------
type AttackState struct {
	Parent *Unit

	IsDoBullet    bool    //是否创建子弹
	StartTime     float64 //开始的时间
	OneAttackTime float64 //一次攻击所需的时间
	IsDone        bool    //是否完成
	AttackTarget  *Unit   //攻击目标
}

func NewAttackState(p *Unit) *AttackState {
	//log.Info(" NewAttackState")
	re := &AttackState{}
	re.Parent = p
	re.OnStart()
	return re
}

//检查状态变换
func (this *AttackState) OnTransform() {

	//攻击完成
	if this.IsDone == true {
		this.OnEnd()
		this.Parent.SetState(NewIdleState(this.Parent))

		//log.Info(" AttackState done%f", utils.GetCurTimeOfSecond())
		return
	}

	//攻击命令
	if this.Parent.HaveAttackCmd() {
		////有攻击指令 却脱离攻击范围内
		if this.Parent.IsOutAttackRangeBuffer(this.AttackTarget) {

			this.OnEnd()
			this.Parent.SetState(NewMoveState(this.Parent))
			return
		}
		//目标不能被攻击
		if this.AttackTarget.IsCanBeAttack() == false {
			this.OnEnd()
			this.Parent.SetState(NewIdleState(this.Parent))
			return
		}
	} else {
		//没有攻击命令 可以移动
		if this.Parent.HaveMoveCmd() && this.Parent.GetCanMove() {
			this.OnEnd()
			this.Parent.SetState(NewMoveState(this.Parent))
			return
		} else {
			//没有攻击命令 不能移动
			this.OnEnd()
			this.Parent.SetState(NewIdleState(this.Parent))
			return
		}
	}

}
func (this *AttackState) Update(dt float64) {
	dotime := utils.GetCurTimeOfSecond() - this.StartTime
	if this.IsDoBullet == false {
		//判断攻击前摇是否完成
		if dotime/this.OneAttackTime >= float64(this.Parent.AttackAnimotionPoint) {
			//创建子弹

			this.IsDoBullet = true
		}
	}

	if dotime/this.OneAttackTime >= 1 {
		this.IsDone = true
	}

}
func (this *AttackState) OnEnd() {
	//log.Info(" AttackState end%f", utils.GetCurTimeOfSecond())
}
func (this *AttackState) OnStart() {
	this.Parent.SetAnimotorState(3)
	this.AttackTarget = this.Parent.AttackCmdDataTarget

	if this.AttackTarget != nil {
		this.Parent.SetDirection(vec2d.Sub(this.AttackTarget.Body.Position, this.Parent.Body.Position))
	}

	//log.Info(" AttackState start%f", utils.GetCurTimeOfSecond())

	this.StartTime = utils.GetCurTimeOfSecond()
	this.IsDoBullet = false
	this.IsDone = false
	this.OneAttackTime = float64(this.Parent.GetOneAttackTime())
}

////------------------------------吟唱状态--------------(玩家使用有吟唱时间的道具或者技能  或者攻击)-----------
//type ChantState struct {
//	Parent *Unit
//}

//func NewChantState(p *Unit) *IdleState {
//	log.Info(" NewChantState")
//	re := &ChantState{}
//	re.Parent = p
//	re.OnStart()
//	return re
//}

////检查状态变换
//func (this *ChantState) OnTransform() {

//	//攻击命令
//	if this.Parent.HaveAttackCmd() {
//		//在攻击范围内
//		if this.Parent.IsInAttackRange(this.Parent.AttackCmdDataTarget) {

//		} else {
//			//有攻击指令 却不在攻击范围内
//			this.OnEnd()
//			this.Parent.SetState(NewMoveState(this.Parent))
//			return
//		}
//	}

//	if this.Parent.HaveMoveCmd() && this.Parent.GetCanMove() {
//		this.OnEnd()
//		this.Parent.SetState(NewMoveState(this.Parent))
//		return
//	}

//}
//func (this *ChantState) Update(dt float64) {
//	this.Parent.SetAnimotorState(1)
//}
//func (this *ChantState) OnEnd() {

//}
//func (this *ChantState) OnStart() {

//}
