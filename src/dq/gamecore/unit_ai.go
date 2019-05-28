package gamecore

import (
	//"dq/log"
	"dq/protobuf"
	"dq/utils"
	//"dq/vec2d"
)

type UnitAI interface {
	Update(dt float64)
	OnEnd()
	OnStart()
	//添加仇人的仇恨值
	AddEnemies(p *Unit, value int32)
	//清除所有仇恨值
	ClearEnemies()
}

//仇人
type Enemies struct {
	Target    *Unit   //单位
	Value     int32   //仇恨值
	FreshTime float64 //最近刷新时间
}

//------------------------------普通AI-------------------------
type NormalAI struct {
	Parent        *Unit
	AllEnemies    map[*Unit]*Enemies
	LastCheckTime float64 //上次检查敌人时间

	AttackTarget *Unit //当前的攻击目标

	//BigEnemies *Enemies //仇恨值最高的仇人
}

func NewNormalAI(p *Unit) *NormalAI {
	//log.Info(" NewNormalAI")
	re := &NormalAI{}
	re.Parent = p
	re.OnStart()
	return re
}

func (this *NormalAI) Update(dt float64) {
	//1秒钟更新1次
	if utils.GetCurTimeOfSecond()-this.LastCheckTime < 1 {
		return
	}

	//更新仇恨列表(5秒)
	this.UpdateEnemies()

	//通过仇恨值攻击目标
	bigEnemies := this.GetBigEnemies()
	if bigEnemies != nil {
		this.AttackTarget = bigEnemies.Target
		this.CreateAttackCmd(bigEnemies.Target)
		return
	}
	//获取最近的敌人
	nearestEnemies := this.GetNearestEnemies()
	if nearestEnemies != nil {
		this.AttackTarget = nearestEnemies
		this.CreateAttackCmd(nearestEnemies)
		return
	}

	//脱离 自动攻击取消追击范围
	if this.Parent.IsOutAutoAttackTraceOutRange(this.AttackTarget) == true {
		this.Parent.StopAttackCmd()
	}

}

func (this *NormalAI) OnEnd() {

}
func (this *NormalAI) OnStart() {
	this.ClearEnemies()
	this.LastCheckTime = 0

}

//创建攻击命令
func (this *NormalAI) CreateAttackCmd(target *Unit) {
	//创建攻击仇人的命令
	data := &protomsg.CS_PlayerAttack{}
	data.IDs = make([]int32, 0)
	data.IDs = append(data.IDs, this.Parent.ID)
	data.TargetUnitID = target.ID
	this.Parent.AttackCmd(data)
}

//更新仇恨列表(5秒)
func (this *NormalAI) UpdateEnemies() {

	for k, v := range this.AllEnemies {

		if utils.GetCurTimeOfSecond()-v.FreshTime > 5 {
			delete(this.AllEnemies, k)
		}

		//if v.Target.InScene.FindUnitByID(v.Target.ID) == nil {
		if v.Target.IsDisappear() || this.Parent.CanSeeTarget(v.Target) == false {
			delete(this.AllEnemies, k)
		}
	}
}

//获取仇恨值最高的目标
func (this *NormalAI) GetBigEnemies() *Enemies {
	var bigEnemies *Enemies = nil
	for _, v := range this.AllEnemies {
		//得到仇恨值最高的仇人
		if bigEnemies == nil {
			bigEnemies = v
		} else {
			if v.Value >= bigEnemies.Value {
				bigEnemies = v
			}
		}

	}
	return bigEnemies
}

//获取最近的敌人
func (this *NormalAI) GetNearestEnemies() *Unit {
	//通过自动攻击范围来攻击目标
	if this.Parent.InScene != nil {
		units := this.Parent.InScene.FindVisibleUnits(this.Parent)
		if len(units) <= 0 {
			return nil
		}
		my := this.Parent

		mindis := 10.0
		var minUnit *Unit = nil
		for _, v := range units {
			//判断阵营 攻击模式 是否死亡  和 是否能被攻击
			if my.CheckAttackEnable2Target(v) {
				//获取目标离本单位自动攻击范围的距离
				dis := my.GetDistanseOfAutoAttackRange(v)
				//判断在自动攻击范围内
				if dis <= 0 {
					if dis < mindis {
						mindis = dis
						minUnit = v
					}
				}
			}
		}

		return minUnit
	}
	return nil
}

func (this *NormalAI) AddEnemies(p *Unit, value int32) {

	one := this.AllEnemies[p]
	if one == nil {
		one = &Enemies{}
		one.Target = p
		this.AllEnemies[p] = one
	}
	one.FreshTime = utils.GetCurTimeOfSecond()
	one.Value += value
}

func (this *NormalAI) ClearEnemies() {
	this.AllEnemies = make(map[*Unit]*Enemies)
	//this.AttackTarget = nil
}
