package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"strings"
)

var BulletID int32 = 100

func GetBulletID() int32 {
	BulletID++
	if BulletID >= 100000000 {
		BulletID = 100
	}
	return BulletID
}

//子弹的影响范围
type BulletRange struct {
	RangeType int32   //范围类型 1:单体 2:圆  3:扇形 4:矩形
	Radius    float32 //半径(圆和扇形)
	Radian    float32 //弧度(扇形)
}

//伤害信息
type HurtInfo struct {
	HurtType  int32 //伤害类型 (1:物理伤害 2:魔法伤害 3:纯粹伤害)
	HurtValue int32 //伤害值
}

//buff信息
type BuffInfo struct {
	Buff      string //buff
	BuffLevel int32  //buff等级
}
type HaloInfo struct {
	Halo      string //buff
	HaloLevel int32  //buff等级
}

type BulletCallUnitInfo struct {
	conf.CallUnitInfo
	CallUnitInfoSkillLevel int32 //技能等级 影响buff和halo等级
}

//技能无视闪避  普工攻击要计算闪避
type Bullet struct {
	ID       int32
	SrcUnit  *Unit         //施法单位
	DestUnit *Unit         //目标单位
	DestPos  vec2d.Vector3 //目的地位置

	Position      vec2d.Vector3 //子弹的当前位置
	StartPosition vec2d.Vector3 //子弹的初始位置

	UnitTargetTeam int32 //目标单位关系 1:友方  2:敌方 3:友方敌方都行

	ModeType string  //子弹模型
	Speed    float32 //子弹速度
	MoveType int32   //移动类型 1:瞬间移动  2:直线移动

	State     int32 //子弹状态(1:创建,2:移动,3:到达后计算结果(伤害和回血) 4:完成 可以删除了)
	NextState int32 //下一帧状态

	StartPosOffsetObj int32 //起始位置参照物   1创建者 2目标 一般情况都为1创建者  像从天而降的闪电就为2目标 的头顶

	SkillID    int32      //技能ID  如果技能ID为 -1表示普通攻击  技能不会miss
	SkillLevel int32      //技能等级
	TargetBuff []BuffInfo //目标buff
	TargetHalo []HaloInfo //目标halo

	NormalHurt HurtInfo   //攻击伤害(以英雄攻击力计算伤害) 值为计算暴击后的值
	OtherHurt  []HurtInfo //其他伤害也就是额外伤害

	HurtRange BulletRange //范围

	Crit                  float32 //暴击倍数
	ClearLevel            int32   //驱散等级
	NoCareDodge           float32 //无视闪避几率
	DoHurtPhysicalAmaorCV float32 //计算伤害时的护甲变化量

	//召唤信息
	BulletCallUnitInfo

	//保存计算伤害后的单位
	HurtUnits      map[int32]*Unit //保存计算伤害后的单位
	IsDoHurtOnMove int32           //在移动的时候也要计算伤害 1:要计算 2:否

	//对目标造成强制移动相关
	ForceMoveTime      float32 //强制移动时间
	ForceMoveSpeedSize float32 //强制移动速度大小
	ForceMoveLevel     int32   //强制移动等级
	ForceMoveBuff      string  //强制移动时的buff 随着移动结束消失

	//加血相关
	AddHPType  int32   //加血类型 0:不加 1:以AddHPValue为固定值 2:以AddHPValue为时间 加单位在此时间内受到的伤害值
	AddHPValue float32 //加血值

	//--------附加攻击特效------

	//发送数据部分
	ClientData    *protomsg.BulletDatas //客户端显示数据
	ClientDataSub *protomsg.BulletDatas //客户端显示差异数据
}

func NewBullet1(src *Unit, dest *Unit) *Bullet {
	re := &Bullet{}
	re.SrcUnit = src
	re.DestUnit = dest
	if dest != nil {
		re.DestPos = dest.GetProjectileEndPos()
	}

	//唯一ID处理
	re.ID = GetBulletID()

	re.Init()

	return re
}
func NewBullet2(src *Unit, pos vec2d.Vec2) *Bullet {
	re := &Bullet{}
	re.SrcUnit = src
	re.DestUnit = nil
	re.DestPos = vec2d.Vector3{pos.X, pos.Y, 0}
	//唯一ID处理
	re.ID = GetBulletID()

	re.Init()

	return re
}
func (this *Bullet) Init() {
	this.Speed = 5
	this.State = 1
	this.NextState = 1
	this.StartPosOffsetObj = 1

	this.SkillID = -1 //普通攻击
	this.SkillLevel = 1

	this.NormalHurt.HurtType = 1 //物理伤害
	this.NormalHurt.HurtValue = 0

	this.OtherHurt = make([]HurtInfo, 0)
	this.TargetBuff = make([]BuffInfo, 0)
	this.TargetHalo = make([]HaloInfo, 0)
	this.HurtUnits = make(map[int32]*Unit)
	this.IsDoHurtOnMove = 2
	this.UnitTargetTeam = 2

	this.HurtRange.RangeType = 1 //单体攻击范围
	this.MoveType = 1            //瞬间移动

	this.Crit = 1
	this.ClearLevel = 0
	this.NoCareDodge = 0
	this.DoHurtPhysicalAmaorCV = 0

	this.AddHPType = 0
	this.AddHPValue = 0

	this.SetForceMove(0, 0, 0, "")
}

//设置加血信息
func (this *Bullet) SetAddHP(hptype int32, val float32) {
	this.AddHPType = hptype
	this.AddHPValue = val
}

//增加无视闪避几率
func (this *Bullet) AddNoCareDodge(val float32) {
	this.NoCareDodge += val
}

//设置削弱护甲
func (this *Bullet) AddDoHurtPhysicalAmaorCV(val int32) {
	if val == -10000 {
		if this.DestUnit == nil || this.DestUnit.IsDisappear() {
			return
		}

		this.DoHurtPhysicalAmaorCV += 0 - this.DestUnit.GetBasePhysicalAmaor()
	} else {
		this.DoHurtPhysicalAmaorCV += float32(val)
	}
}

//设置强制移动相关
func (this *Bullet) SetForceMove(time float32, speedsize float32, level int32, buff string) {
	//对目标造成强制移动相关
	this.ForceMoveTime = time           //强制移动时间
	this.ForceMoveSpeedSize = speedsize //强制移动速度大小
	this.ForceMoveLevel = level         //强制移动等级
	this.ForceMoveBuff = buff

}

//设置范围
func (this *Bullet) SetRange(r float32) {
	if r <= 0 {
		this.HurtRange.RangeType = 1 //单体攻击范围
	} else {
		this.HurtRange.RangeType = 2 //圆
		this.HurtRange.Radius = r
	}
}

//设置暴击倍数
func (this *Bullet) SetCrit(crit float32) {
	if crit > this.Crit {
		this.Crit = crit
	}
}

//设置弹道
func (this *Bullet) SetProjectileMode(modetype string, speed float32) {
	this.ModeType = modetype
	this.Speed = speed
	if this.Speed <= 0 || this.Speed >= 1000000 {
		this.MoveType = 1
	} else {
		this.MoveType = 2
	}
}

//设置普通攻击伤害百分比
func (this *Bullet) SetNormalHurtRatio(ratio float32) {
	if this.SrcUnit == nil {
		return
	}
	//物理伤害
	this.NormalHurt.HurtType = 1
	this.NormalHurt.HurtValue = int32(float32(this.SrcUnit.Attack) * ratio)

	//log.Info("---SetNormalHurtRatio---%d", this.NormalHurt.HurtValue)

}

////计算初始位置
func (this *Bullet) OnCreate() {

	//起始相对位置为创建者的时候
	if this.StartPosOffsetObj == 1 {
		if this.SrcUnit == nil {
			this.Done()
			return
		}

		this.Position = this.SrcUnit.GetProjectileStartPos()
		//开始位置
		this.StartPosition = this.Position.Clone()

		//		if this.DestUnit != nil {
		//			this.DestPos = this.DestUnit.GetProjectileEndPos()
		//		}

	}

	//如果为瞬间移动则直接计算伤害
	if this.MoveType == 1 {
		this.NextState = 3
	} else {
		this.NextState = 2
	}

}

//移动
func (this *Bullet) DoMove(dt float32) {
	//计算终点位置
	if this.DestUnit != nil {
		if this.DestUnit.IsDisappear() == false && this.SrcUnit.CanSeeTarget(this.DestUnit) {
			this.DestPos = this.DestUnit.GetProjectileEndPos()
		} else {
			this.DestUnit = nil
		}
	}

	//
	movedis := float64(this.Speed * dt)
	dis := (vec2d.GetDistance3(this.Position, this.DestPos))
	//到达终点
	if dis <= movedis {
		this.Position = this.DestPos
		this.NextState = 3
		return
	} else {
		//t1 := vec2d.Sub3(this.DestPos, this.Position)
		movevec3 := vec2d.Normalize3(vec2d.Sub3(this.DestPos, this.Position))
		movevec3.Multiply(movedis)
		//movevec3 := vec2d.Normalize3(vec2d.Sub3(this.DestPos, this.Position)).Multiply(movedis)
		this.Position.Add(movevec3)
	}
}

//增加目标buff
func (this *Bullet) AddTargetBuff(buff string, level int32) {
	this.TargetBuff = append(this.TargetBuff, BuffInfo{buff, level})
}

//增加目标buff
func (this *Bullet) AddTargetHalo(buff string, level int32) {
	this.TargetHalo = append(this.TargetHalo, HaloInfo{buff, level})
	log.Info("---AddTargetHalo:%s   %d", buff, level)
}

//处理召唤
func (this *Bullet) DoCallUnit() {
	//	type CallUnitInfo struct {
	//	//召唤相关
	//	CallUnitCount     int32   //召唤数量 0表示没有召唤
	//	CallUnitTypeID    int32   //召唤出来的单位 类型ID 0表示当前召唤者 -1表示目标对象 其他类型id对应其他单位
	//	CallUnitBuff      string  //召唤出来的单位携带额外buff
	//	CallUnitHalo      string  //召唤出来的单位携带额外halo
	//	CallUnitOffsetPos float32 //召唤出来的单位在目标位置的随机偏移位置
	//	CallUnitAliveTime float32 //召唤单位的生存时间
	//}

	if this.SrcUnit == nil || this.SrcUnit.IsDisappear() {
		return
	}
	scene := this.SrcUnit.InScene
	for i := int32(0); i < this.CallUnitCount; i++ {

		if this.CallUnitTypeID > 0 {
			unit := CreateUnit(this.SrcUnit.InScene, this.CallUnitTypeID)
			if unit == nil {
				continue
			}
			//设置移动核心body
			//pos := vec2d.Vec2{float64(0), float64(0)}
			//r := vec2d.Vec2{unit.CollisionR, unit.CollisionR}
			//unit.Body = scene.MoveCore.CreateBody(pos, r, 0, 1)
			p1 := vec2d.Vec2{this.DestPos.X + float64(utils.GetRandomFloat(this.CallUnitOffsetPos)),
				this.DestPos.Z + float64(utils.GetRandomFloat(this.CallUnitOffsetPos))}

			log.Info("------------pos:%v", p1)
			unit.InitPosition = p1
			//unit.Body.BlinkToPos(p1)
			//
			unit.Camp = this.SrcUnit.Camp

			//buff
			unit.AddBuffFromStr(this.CallUnitBuff, this.CallUnitInfoSkillLevel, unit)
			unit.AddHaloFromStr(this.CallUnitHalo, this.CallUnitInfoSkillLevel, nil)

			scene.NextAddUnit.Set(unit.ID, unit)
		}
	}
}
func (this *Bullet) GetPosition2D() vec2d.Vec2 {
	return vec2d.Vec2{this.Position.X, this.Position.Y}
}

//对单位造成伤害 只计算一次
func (this *Bullet) HurtUnit(unit *Unit) {
	//log.Info("222222222")
	if unit == nil {
		return
	}
	if _, ok := this.HurtUnits[unit.ID]; ok {
		return
	}
	//log.Info("33333333")

	this.HurtUnits[unit.ID] = unit

	//对目标加血
	if this.AddHPType != 0 {
		unit.DoAddHP(this.AddHPType, this.AddHPValue)
	}

	//伤害
	ismiss, hurtvalue := unit.BeAttacked(this)

	//强制移动
	if this.ForceMoveTime > 0 {
		if this.HurtRange.RangeType == 1 {

			dir := vec2d.Sub(unit.Body.Position, vec2d.Vec2{this.StartPosition.X, this.StartPosition.Y})
			dir.Normalize()
			dir.MulToFloat64(float64(this.ForceMoveSpeedSize))
			unit.SetForceMove(this.ForceMoveTime, dir, this.ForceMoveLevel)
		} else {
			dir := vec2d.Sub(unit.Body.Position, this.GetPosition2D())
			dir.Normalize()
			dir.MulToFloat64(float64(this.ForceMoveSpeedSize))
			unit.SetForceMove(this.ForceMoveTime, dir, this.ForceMoveLevel)
		}
		//更改buff时间
		if len(this.ForceMoveBuff) > 0 {
			buffs := unit.AddBuffFromStr(this.ForceMoveBuff, this.SkillLevel, this.SrcUnit)
			for _, v := range buffs {
				v.RemainTime = this.ForceMoveTime
				v.Time = this.ForceMoveTime
			}
		}
	}
	//	//光环
	//	if len(this.TargetHalo) > 0 {
	//		if this.DestUnit != nil {
	//			if this.DestUnit.IsDisappear() == false {
	//				for _, v := range this.TargetHalo {
	//					this.DestUnit.AddHaloFromStr(v.Halo, v.HaloLevel, &vec2d.Vec2{this.DestPos.X, this.DestPos.Y})
	//				}
	//			}
	//		} else {
	//			if this.SrcUnit != nil && this.SrcUnit.IsDisappear() == false {
	//				for _, v := range this.TargetHalo {
	//					this.SrcUnit.AddHaloFromStr(v.Halo, v.HaloLevel, &vec2d.Vec2{this.DestPos.X, this.DestPos.Y})
	//				}
	//			}
	//		}
	//	}
	//小于0 表示被miss
	if ismiss == false {
		//驱散buff
		unit.ClearBuffForTarget(this.SrcUnit, this.ClearLevel)
		//buff
		for _, v := range this.TargetBuff {
			unit.AddBuffFromStr(v.Buff, v.BuffLevel, this.SrcUnit)
		}
		if this.SrcUnit == nil || this.SrcUnit.MyPlayer == nil {
			return
		}
		//为了显示 玩家造成的伤害
		mph := &protomsg.MsgPlayerHurt{HurtUnitID: unit.ID, HurtAllValue: hurtvalue}
		if this.Crit > 1 {
			mph.IsCrit = 1
		}
		this.SrcUnit.MyPlayer.AddHurtValue(mph)
	}

}
func (this *Bullet) DoHalo() {
	//光环
	if len(this.TargetHalo) > 0 {
		if this.DestUnit != nil {
			if this.DestUnit.IsDisappear() == false {
				for _, v := range this.TargetHalo {
					this.DestUnit.AddHaloFromStr(v.Halo, v.HaloLevel, &vec2d.Vec2{this.DestPos.X, this.DestPos.Y})
				}
			}
		} else {
			if this.SrcUnit != nil && this.SrcUnit.IsDisappear() == false {
				for _, v := range this.TargetHalo {
					this.SrcUnit.AddHaloFromStr(v.Halo, v.HaloLevel, &vec2d.Vec2{this.DestPos.X, this.DestPos.Y})
				}
			}
		}
	}
}

//计算伤害
func (this *Bullet) DoHurt() {
	//获取到受伤害的单位 (狂战斧攻击特效也会影响单位数量 )
	//log.Info("111111111111")
	if this.HurtRange.RangeType == 1 {
		this.DoHalo()
		//单体范围
		if this.DestUnit == nil {
			return
		}
		this.HurtUnit(this.DestUnit)
	} else if this.HurtRange.RangeType == 2 {
		if this.SrcUnit == nil || this.SrcUnit.IsDisappear() {
			return
		}
		this.DoHalo()
		allunit := this.SrcUnit.InScene.FindVisibleUnitsByPos(this.GetPosition2D())
		for _, v := range allunit {
			if v.IsDisappear() {
				continue
			}
			//UnitTargetTeam      int32   //目标单位关系 1:友方  2:敌方 3:友方敌方都行
			if this.UnitTargetTeam == 1 && this.SrcUnit.CheckIsEnemy(v) == true {
				continue
			}
			if this.UnitTargetTeam == 2 && this.SrcUnit.CheckIsEnemy(v) == false {
				continue
			}
			//检测是否在范围内
			if v.Body == nil || this.HurtRange.Radius <= 0 {
				continue
			}
			dis := float32(vec2d.Distanse(this.GetPosition2D(), v.Body.Position))
			//log.Info("-----------------dis:%f", dis)
			if dis <= this.HurtRange.Radius {
				this.HurtUnit(v)
			}

		}
	}

}
func (this *Bullet) AddOtherHurt(hurtinfo HurtInfo) {

	this.OtherHurt = append(this.OtherHurt, hurtinfo)
}

//伤害类型 (1:物理伤害 2:魔法伤害 3:纯粹伤害)
//计算攻击力  根据攻击类型
func (this *Bullet) GetAttackOfType(hurttype int32) int32 {

	val := int32(0)
	if this.NormalHurt.HurtType == hurttype {
		val += this.NormalHurt.HurtValue
	}
	for _, v := range this.OtherHurt {
		if v.HurtType == hurttype {
			val += v.HurtValue
		}
	}
	//计算暴击
	if hurttype == 1 && this.Crit > 1 {
		val = int32(float32(val) * this.Crit)
	}

	return val
}

//计算结果
func (this *Bullet) CalResult() {

	//处理召唤
	this.DoCallUnit()
	//计算伤害
	this.DoHurt()

	//等待删除
	this.Done()
}

//(1:创建,2:移动,3:到达后计算伤害 4:完成 可以删除了)
func (this *Bullet) Update(dt float32) {

	this.State = this.NextState

	if this.State == 1 {
		this.OnCreate()
	} else if this.State == 2 {
		this.DoMove(dt)
		if this.IsDoHurtOnMove == 1 {
			this.DoHurt()
		}
	} else if this.State == 3 {
		this.CalResult()
	} else if this.State == 4 {
		//log.Info("---Bullet update---%d", this.State)
	}
	//log.Info("---Bullet update---%d", this.State)
}

//子弹done
func (this *Bullet) Done() {
	this.NextState = 4
}
func (this *Bullet) IsDone() bool {
	if this.State == 4 {
		return true
	}
	return false
}

//客户端是否要显示
func (this *Bullet) ClientIsShow() bool {
	if len(this.ModeType) > 0 {
		return true
	}
	return false
}

//刷新客户端显示数据
func (this *Bullet) FreshClientData() {
	if this.ClientIsShow() == false {
		return
	}

	if this.ClientData == nil {
		this.ClientData = &protomsg.BulletDatas{}
	}

	this.ClientData.ID = this.ID

	this.ClientData.X = float32(this.Position.X)
	this.ClientData.Y = float32(this.Position.Y)
	this.ClientData.Z = float32(this.Position.Z)
	this.ClientData.StartX = float32(this.StartPosition.X)
	this.ClientData.StartY = float32(this.StartPosition.Y)
	this.ClientData.StartZ = float32(this.StartPosition.Z)
	this.ClientData.EndX = float32(this.DestPos.X)
	this.ClientData.EndY = float32(this.DestPos.Y)
	this.ClientData.EndZ = float32(this.DestPos.Z)

	this.ClientData.State = this.State
	this.ClientData.ModeType = this.ModeType

}

//刷新客户端显示差异数据
func (this *Bullet) FreshClientDataSub() {
	if this.ClientIsShow() == false {
		return
	}

	if this.ClientDataSub == nil {
		this.ClientDataSub = &protomsg.BulletDatas{}
	}
	if this.ClientData == nil {
		this.FreshClientData()
		*this.ClientDataSub = *this.ClientData
		return
	}
	//
	//字符串部分
	if strings.Compare(this.ModeType, this.ClientData.ModeType) != 0 {
		this.ClientDataSub.ModeType = this.ModeType
	} else {
		this.ClientDataSub.ModeType = ""
	}

	//当前数据与上一次数据对比 相减 数值部分
	this.ClientDataSub.X = float32(this.Position.X) - this.ClientData.X
	this.ClientDataSub.Y = float32(this.Position.Y) - this.ClientData.Y
	this.ClientDataSub.Z = float32(this.Position.Z) - this.ClientData.Z
	this.ClientDataSub.StartX = float32(this.StartPosition.X) - this.ClientData.StartX
	this.ClientDataSub.StartY = float32(this.StartPosition.Y) - this.ClientData.StartY
	this.ClientDataSub.StartZ = float32(this.StartPosition.Z) - this.ClientData.StartZ
	this.ClientDataSub.EndX = float32(this.DestPos.X) - this.ClientData.EndX
	this.ClientDataSub.EndY = float32(this.DestPos.Y) - this.ClientData.EndY
	this.ClientDataSub.EndZ = float32(this.DestPos.Z) - this.ClientData.EndZ

	this.ClientDataSub.State = this.State - this.ClientData.State

}
