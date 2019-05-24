package gamecore

import (
	"dq/log"
	//"dq/protobuf"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
)

type UnitState interface {
	OnTransform()
	Update(dt float64)
	OnEnd()
	OnStart()
	GetParent() *Unit
	GetStateID() int32 //1:idle 2:move 3:attack 4:death
}

//全局状态切换检查
func GTransform(this UnitState) bool {

	parent := this.GetParent()
	if parent == nil {
		return false
	}
	//死亡
	if parent.IsDeath == 1 && this.GetStateID() != 4 {
		this.OnEnd()
		parent.SetState(NewDeathState(parent))
		return true
	}
	return false
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

func (this *IdleState) GetParent() *Unit {
	return this.Parent
}
func (this *IdleState) GetStateID() int32 {
	return 1
}

//检查状态变换
func (this *IdleState) OnTransform() {

	//全局状态切换检查
	if GTransform(this) {
		return
	}
	//技能命令
	if this.Parent.HaveSkillCmd() {
		//在技能范围内
		if this.Parent.IsInSkillRange(this.Parent.SkillCmdData) {

			this.OnEnd()
			this.Parent.SetState(NewChantState(this.Parent))

		} else {
			//有攻击指令 却不在攻击范围内
			this.OnEnd()
			this.Parent.SetState(NewMoveState(this.Parent))
		}
		return
	}

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

func (this *MoveState) GetParent() *Unit {
	return this.Parent
}
func (this *MoveState) GetStateID() int32 {
	return 2
}

//检查状态变换
func (this *MoveState) OnTransform() {

	//全局状态切换检查
	if GTransform(this) {
		return
	}
	//技能命令
	if this.Parent.HaveSkillCmd() {
		//在技能范围内
		if this.Parent.IsInSkillRange(this.Parent.SkillCmdData) {

			this.OnEnd()
			this.Parent.SetState(NewChantState(this.Parent))
			return

		}
		return
	}

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
	//先检查技能对象
	if this.Parent.HaveSkillCmd() {
		//上次寻路的目标单位和本次相同则在1S内 不再寻路
		if this.LastFindPathTarget != nil {
			if this.LastFindPathTarget.ID == this.Parent.SkillCmdData.TargetUnitID {
				if utils.GetCurTimeOfSecond()-this.LastFindPathTargetTime < 1 {
					return
				}
			}
		}

		//有目标单位
		if this.Parent.SkillCmdData.TargetUnitID > 0 {
			targetunit := this.Parent.InScene.FindUnitByID(this.Parent.SkillCmdData.TargetUnitID)
			if targetunit != nil {
				this.Parent.Body.SetTarget(targetunit.Body.Position)
				this.LastFindPathTarget = targetunit
				this.LastFindPathTargetTime = utils.GetCurTimeOfSecond()
			}
		} else {
			targetpos := vec2d.Vec2{X: float64(this.Parent.SkillCmdData.X), Y: float64(this.Parent.SkillCmdData.Y)}
			this.Parent.Body.SetTarget(targetpos)
		}
		return
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

func (this *AttackState) GetParent() *Unit {
	return this.Parent
}
func (this *AttackState) GetStateID() int32 {
	return 3
}

//检查状态变换
func (this *AttackState) OnTransform() {

	//全局状态切换检查
	if GTransform(this) {
		return
	}

	//技能命令
	if this.Parent.HaveSkillCmd() {
		//在技能范围内
		if this.Parent.IsInSkillRange(this.Parent.SkillCmdData) {

			this.OnEnd()
			this.Parent.SetState(NewChantState(this.Parent))
			return

		} else {
			//有攻击指令 却不在攻击范围内
			this.OnEnd()
			this.Parent.SetState(NewMoveState(this.Parent))
			return
		}
	}

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
		//切换目标
		if this.AttackTarget != this.Parent.AttackCmdDataTarget {
			this.OnEnd()
			this.Parent.SetState(NewIdleState(this.Parent))
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

			b := NewBullet1(this.Parent, this.AttackTarget)
			b.SetNormalHurtRatio(1)
			b.SetProjectileMode(this.Parent.ProjectileMode, this.Parent.ProjectileSpeed)
			this.Parent.AddBullet(b)

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

	this.StartTime = utils.GetCurTimeOfSecond()
	this.IsDoBullet = false
	this.IsDone = false
	this.OneAttackTime = float64(this.Parent.GetOneAttackTime())

	if this.Parent.UnitType == 1 {
		log.Info(" Attacktime%f   speed:%f", this.OneAttackTime, this.Parent.AttackSpeed)
	}
}

//------------------------------死亡状态-------------------------
type DeathState struct {
	Parent *Unit

	StartTime float64 //开始的时间
}

func NewDeathState(p *Unit) *DeathState {
	//log.Info(" NewDeathState")
	re := &DeathState{}
	re.Parent = p
	re.OnStart()
	return re
}

func (this *DeathState) GetParent() *Unit {
	return this.Parent
}
func (this *DeathState) GetStateID() int32 {
	return 4
}

//检查状态变换
func (this *DeathState) OnTransform() {
	//全局状态切换检查
	if GTransform(this) {
		return
	}
	//复活
	if this.Parent.IsDeath != 1 {
		this.OnEnd()
		this.Parent.SetState(NewIdleState(this.Parent))
	}

}
func (this *DeathState) Update(dt float64) {
	//this.Parent.SetAnimotorState(1)
	dotime := utils.GetCurTimeOfSecond() - this.StartTime
	if dotime >= 2 {
		this.Parent.SetAnimotorState(5)
	}
	if dotime >= 4 {
		this.Parent.InScene.RemoveUnit(this.Parent)
	}
}
func (this *DeathState) OnEnd() {

}
func (this *DeathState) OnStart() {
	this.Parent.SetAnimotorState(4)

	this.StartTime = utils.GetCurTimeOfSecond()
}

//---------------------------吟唱状态--------------玩家使用有吟唱时间的道具或者技能---------------
type ChantState struct {
	Parent *Unit

	IsDoBullet   bool    //是否创建子弹
	StartTime    float64 //开始的时间
	OneChantTime float64 //一次吟唱所需的时间
	IsDone       bool    //是否完成
	AttackTarget *Unit   //攻击目标

	ChantData *protomsg.CS_PlayerSkill //技能数据
	CastPoint float32

	StartTargetPos vec2d.Vec2 //开始时的 目标位置
}

func NewChantState(p *Unit) *ChantState {
	//log.Info(" NewAttackState")
	re := &ChantState{}
	re.Parent = p
	re.OnStart()
	return re
}

func (this *ChantState) GetParent() *Unit {
	return this.Parent
}
func (this *ChantState) GetStateID() int32 {
	return 5
}

//检查状态变换
func (this *ChantState) OnTransform() {

	//全局状态切换检查
	if GTransform(this) {
		return
	}

	//吟唱完成
	if this.IsDone == true {
		this.OnEnd()
		this.Parent.SetState(NewIdleState(this.Parent))

		//log.Info(" AttackState done%f", utils.GetCurTimeOfSecond())
		return
	}

	//技能命令
	if this.Parent.HaveSkillCmd() {

		//切换目标
		if this.ChantData != this.Parent.SkillCmdData {
			this.OnEnd()
			this.Parent.SetState(NewIdleState(this.Parent))
		}
	} else {
		//没有攻击命令 不能移动
		this.OnEnd()
		this.Parent.SetState(NewIdleState(this.Parent))
		return
	}

}
func (this *ChantState) Update(dt float64) {
	dotime := utils.GetCurTimeOfSecond() - this.StartTime
	if this.IsDoBullet == false {
		//判断攻击前摇是否完成
		if dotime/this.OneChantTime >= float64(this.CastPoint) {
			//创建子弹

			//			b := NewBullet1(this.Parent, this.AttackTarget)
			//			b.SetNormalHurtRatio(1)
			//			b.SetProjectileMode(this.Parent.ProjectileMode, this.Parent.ProjectileSpeed)
			//			this.Parent.CreateBullet(b)
			this.Parent.DoSkill(this.ChantData, this.StartTargetPos)

			//this.Parent.Body.BlinkToPos(this.StartTargetPos)

			this.IsDoBullet = true

		}
	}

	if dotime/this.OneChantTime >= 1 {
		this.IsDone = true
	}

}
func (this *ChantState) OnEnd() {
	log.Info(" ChantState end%f", utils.GetCurTimeOfSecond())
}
func (this *ChantState) OnStart() {
	this.Parent.SetAnimotorState(10)

	this.ChantData = this.Parent.SkillCmdData

	skilldata, ok := this.Parent.Skills[this.Parent.SkillCmdData.SkillID]
	if ok == false {
		this.IsDone = true
		return
	}
	//--转向处理--
	if skilldata.CastTargetType == 2 {
		target := this.Parent.InScene.FindUnitByID(this.Parent.SkillCmdData.TargetUnitID)
		if target == nil {
			this.IsDone = true
			return
		}
		this.StartTargetPos = target.Body.Position
		this.Parent.SetDirection(vec2d.Sub(target.Body.Position, this.Parent.Body.Position))
	} else if skilldata.CastTargetType == 2 {
		targetpos := vec2d.Vec2{X: float64(this.Parent.SkillCmdData.X), Y: float64(this.Parent.SkillCmdData.Y)}
		this.StartTargetPos = targetpos
		this.Parent.SetDirection(vec2d.Sub(targetpos, this.Parent.Body.Position))
	}

	log.Info(" ChantState start%f", utils.GetCurTimeOfSecond())

	this.StartTime = utils.GetCurTimeOfSecond()
	this.IsDoBullet = false
	this.IsDone = false
	this.OneChantTime = float64(skilldata.CastTime)
	this.CastPoint = skilldata.CastPoint
}
