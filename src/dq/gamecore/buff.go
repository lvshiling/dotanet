package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"strconv"
)

type Buff struct {
	conf.BuffData //技能数据

	Parent            *Unit   //载体
	CastUnit          *Unit   //施加buff的单位
	Level             int32   //当前等级
	RemainTime        float32 //剩余时间
	TagNum            int32   //标记数字 拍拍熊被动 大圣被动
	TriggerRemainTime float32 //触发剩余时间 造成伤害之类的

	IsEnd bool //是否已经结束

	IsActive bool //是否生效

	IsUseable     bool  //是否起作用 对特定单位其中有功能
	UseableUnitID int32 //起作用的单位ID

	MagicHurtBlockCurValue int32 //当前吸收的魔法伤害量

	LastPos          vec2d.Vec2 //载体上次计算的位置 血魔大招
	PathHaloLastTime float64    //上次的路径光环时间

	AttackTarget *Unit //攻击目标

	ConnectionType  int32      //是否有连接 0表示没有 1表示有连接点 2表示有连接单位
	ConnectionPoint vec2d.Vec2 //连接点
	ConnectionUnit  *Unit      //连接单位
}

//处理强制攻击
func (this *Buff) UpdateForceAttack() {
	if this.ForceAttackTarget == 1 {
		if this.AttackTarget == nil {
			this.AttackTarget = this.Parent.AttackCmdDataTarget
		}
		//攻击目标
		if this.AttackTarget != nil {
			//创建攻击命令
			data := &protomsg.CS_PlayerAttack{}
			data.IDs = make([]int32, 0)
			data.IDs = append(data.IDs, this.Parent.ID)
			data.TargetUnitID = this.AttackTarget.ID
			this.Parent.AttackCmd(data)
		}
	}
}

//设置牵连
func (this *Buff) SetConnectionPoint(p vec2d.Vec2) {
	this.ConnectionType = 1
	this.ConnectionPoint = p
}

//检查对特定攻击对象起作用
func (this *Buff) FreshUseable(unit *Unit) {

	if this.IsUseableAllocateAttackUnit == 2 {
		this.IsUseable = true
		return
	}

	lastuseable := this.IsUseable
	if this.UseableUnitID <= 0 {
		this.IsUseable = true
	} else {
		if unit == nil || unit.ID != this.UseableUnitID {
			this.IsUseable = false
		} else {
			this.IsUseable = true
		}
	}
	//从不起作用到起作用
	if lastuseable == false && this.IsUseable == true {
		add := &UnitBaseProperty{}
		this.Parent.CalPropertyByBuff(this, add)
		this.Parent.AddBuffProperty(add)
	}
}

func (this *Buff) ExceptionTrigger() {
	if this.Exception <= 0 {
		return
	}
	switch this.Exception {
	case 3: //血魔的大招
		{
			this.HurtValue = 0
			if this.Parent == nil || this.Parent.Body == nil {
				return
			}
			param := utils.GetFloat32FromString3(this.ExceptionParam, ":")
			if len(param) < 1 {
				return
			}

			dis := vec2d.Sub(this.Parent.Body.Position, this.LastPos)
			this.HurtValue = param[0] * float32(dis.Length())
			log.Info("----hurt:%f", this.HurtValue)
			this.LastPos = this.Parent.Body.Position
		}
	case 4: //帕克大招
		{
			//log.Info("----1111111111aaaaaaaaa")
			if this.Parent == nil || this.Parent.Body == nil {
				return
			}
			param := utils.GetFloat32FromString3(this.ExceptionParam, ":")
			if len(param) < 4 {
				return
			}
			hurt := param[1]
			hurttype := param[0]
			distanse := param[2]
			buffid := strconv.Itoa(int(param[3]))

			dis := vec2d.Sub(this.Parent.Body.Position, this.ConnectionPoint)
			if distanse < float32(dis.Length()) {
				//触发
				castunit := this.Parent
				if this.CastUnit != nil {
					castunit = this.CastUnit
				}
				b := NewBullet1(castunit, this.Parent)
				b.SetProjectileMode("", 0)
				b.AddOtherHurt(HurtInfo{HurtType: int32(hurttype), HurtValue: int32(hurt)})
				b.AddTargetBuff(buffid, this.Level)
				if b != nil {
					this.Parent.AddBullet(b)
				}

				//log.Info("----22222222aaaaaaaaa")

				//删除自己
				this.RemainTime = 0
				this.IsEnd = true
			}

		}
	default:
		{

		}
	}
}

//创建路径光环
func (this *Buff) DoCreatePathHalo() {
	castunit := this.Parent
	if this.CastUnit != nil {
		castunit = this.CastUnit
	}
	if castunit == nil || castunit.Body == nil || this.Parent == nil || this.Parent.Body == nil || len(this.PathHalo) <= 0 {
		return
	}

	curtime := utils.GetCurTimeOfSecond()
	//log.Info("-DoCreatePathHalo-:%f  :%f  :%f", curtime, this.PathHaloLastTime, this.PathHaloMinTime)
	if curtime-this.PathHaloLastTime > float64(this.PathHaloMinTime) {
		pos := this.Parent.Body.Position

		castunit.AddHaloFromStr(this.PathHalo, this.Level, &pos)

		this.PathHaloLastTime = curtime
		log.Info("buff DoCreatePathHalo succ :%f", curtime)

	}

}

func (this *Buff) GetHurtValue(src *Unit, dest *Unit) int32 {
	switch this.HurtValueType {
	case 0:
		return int32(this.HurtValue)
	case 1: //受到总伤害比例 的比例
		{
			if dest == nil || dest.IsDisappear() {
				return 0
			}
			return int32(this.HurtValue * (1 - float32(dest.HP)/float32(dest.MAX_HP)) * 100.0)
		}
	case 2: //最大血量百分比
		{
			if dest == nil || dest.IsDisappear() {
				return 0
			}
			return int32(this.HurtValue * float32(dest.MAX_HP))
		}
	case 3: //受到总伤害比例
		{
			if dest == nil || dest.IsDisappear() {
				return 0
			}
			return int32(this.HurtValue * (1 - float32(dest.HP)/float32(dest.MAX_HP)) * float32(dest.MAX_HP))
		}
	}

	return 0

}

//魔法伤害吸收
func (this *Buff) BlockMagicHurt(hurt int32) int32 {
	if this.MagicHurtBlockAllValue > 0 {
		maxblockvalue := this.MagicHurtBlockAllValue - this.MagicHurtBlockCurValue
		log.Info("--BlockMagicHurt-:%d   %d", this.MagicHurtBlockAllValue, this.MagicHurtBlockCurValue)
		if maxblockvalue > 0 {
			if maxblockvalue > hurt {
				this.MagicHurtBlockCurValue = this.MagicHurtBlockCurValue + hurt
				return 0
			} else {
				this.MagicHurtBlockCurValue = this.MagicHurtBlockAllValue
				return hurt - maxblockvalue
			}
		}
	}
	return hurt
}

//更新
func (this *Buff) Update(dt float64) {
	//CD时间减少
	if this.IsActive {
		this.RemainTime -= float32(dt)

		//魔法伤害吸收完成
		if this.MagicHurtBlockAllValue > 0 {
			if this.MagicHurtBlockCurValue >= this.MagicHurtBlockAllValue {
				this.IsEnd = true
			}
		}

		if this.RemainTime <= 0.00001 {
			this.RemainTime = 0
			this.IsEnd = true
		}
		//		if this.TypeID == 70 {
		//			log.Info("buffupdate:%f  TypeID:%d", this.RemainTime, this.TypeID)
		//		}

		this.DoCreatePathHalo()
		this.UpdateForceAttack()

		this.TriggerRemainTime -= float32(dt)
		//log.Info("time:%f  :%f", this.RemainTime, this.TriggerRemainTime)
		//检查是否触发
		if this.TriggerRemainTime <= 0.00001 {
			//重置触发时间
			this.TriggerRemainTime = this.HurtTimeInterval + this.TriggerRemainTime

			if this.Parent != nil && this.Parent.InScene != nil && this.Parent.IsDisappear() == false {
				//创建触发子弹 //伤害类型(1:物理伤害 2:魔法伤害 3:纯粹伤害 4:不造成伤害)
				//创建buff
				if this.HurtType != 4 {
					castunit := this.Parent
					if this.CastUnit != nil {
						castunit = this.CastUnit
					}
					b := NewBullet1(castunit, this.Parent)
					b.SetProjectileMode("", 0)
					this.ExceptionTrigger()
					b.AddOtherHurt(HurtInfo{HurtType: this.HurtType, HurtValue: this.GetHurtValue(castunit, this.Parent)})
					if b != nil {
						this.Parent.AddBullet(b)
					}
				} else {
					this.ExceptionTrigger()
				}
			}
		}

	} else {
		this.ActiveTime -= float32(dt)
		if this.ActiveTime <= 0 {
			this.ActiveTime = 0
			this.IsActive = true
		} else {
			this.IsActive = false
		}
		//log.Info("-----activetime:%f", this.ActiveTime)
	}

}

//创建buf
func NewBuff(typeid int32, level int32, parent *Unit) *Buff {

	buffdata := conf.GetBuffData(typeid, level)
	if buffdata == nil {
		log.Error("NewBuff %d  %d", typeid, level)
		return nil
	}
	buff := &Buff{}
	buff.BuffData = *buffdata
	buff.Level = level
	buff.RemainTime = buffdata.Time
	buff.TagNum = buffdata.InitTagNum
	buff.IsEnd = false

	if buffdata.ActiveTime <= 0 {
		buff.IsActive = true
	} else {
		buff.IsActive = false
	}
	buff.Parent = parent
	if parent != nil && parent.Body != nil {
		buff.LastPos = parent.Body.Position
	}
	buff.TriggerRemainTime = buffdata.HurtTimeInterval
	if buff.IsUseableAllocateAttackUnit == 1 {
		buff.IsUseable = false //是否起作用
	} else {
		buff.IsUseable = true //是否起作用
	}
	buff.UseableUnitID = 0 //起作用的单位ID

	return buff

}
