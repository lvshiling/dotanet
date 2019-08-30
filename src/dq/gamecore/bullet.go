package gamecore

import (
	"dq/conf"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"strconv"
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

//子弹的影响范围 范围内的单位和目标单位受到一样的效果
type BulletRange struct {
	RangeType int32   //范围类型 1:单体 2:圆  3:扇形 4:矩形
	Radius    float32 //半径(圆和扇形)
	Radian    float32 //弧度(扇形)
}

//溅射  溅射范围内的单位受到目标单位伤害效果的一定百分比
type BulletSputtering struct {
	HurtRatio         float32 //伤害百分比
	Radius            float32 //半径(圆和扇形)
	Radian            float32 //弧度(扇形) 360度为圆形
	NoCareMagicImmune int32   //是否无视单位魔法免疫 1:是 2:非
}

type BulletEjection struct {
	EjectionCount    int32           //弹射次数
	EjectionRange    float32         //弹射范围
	EjectionDecay    float32         //弹射衰减
	EjectionedTarget map[int32]*Unit //弹射过的目标 不能重复弹射
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

	ModeType             string  //子弹模型
	Speed                float32 //子弹速度
	MoveType             int32   //移动类型 1:瞬间移动  2:直线移动
	UseUnitProjectilePos int32   //是否使用单位攻击弹道起始点

	State     int32 //子弹状态(1:创建,2:移动,3:到达后计算结果(伤害和回血) 4:完成 可以删除了)
	NextState int32 //下一帧状态

	StartPosOffsetObj int32 //起始位置参照物   1创建者 2目标 一般情况都为1创建者  像从天而降的闪电就为2目标 的头顶

	SkillID    int32      //技能ID  如果技能ID为 -1表示普通攻击  技能不会miss
	SkillLevel int32      //技能等级
	TargetBuff []BuffInfo //目标buff
	TargetHalo []HaloInfo //目标halo

	NormalHurt HurtInfo   //攻击伤害(以英雄攻击力计算伤害) 值为计算暴击后的值
	OtherHurt  []HurtInfo //其他伤害也就是额外伤害

	Sputterings []BulletSputtering //溅射

	Ejection BulletEjection //弹射

	HurtRange BulletRange //范围

	Crit                  float32 //暴击倍数
	ClearLevel            int32   //驱散等级
	NoCareDodge           float32 //无视闪避几率
	DoHurtPhysicalAmaorCV float32 //计算伤害时的护甲变化量

	//召唤信息
	BulletCallUnitInfo

	//保存计算伤害后的单位
	HurtUnits               map[int32]*Unit //保存计算伤害后的单位
	IsDoHurtOnMove          int32           //在移动的时候也要计算伤害 1:要计算 2:否
	EveryDoHurtChangeHurtCR float32         //每对一个目标造成伤害后 伤害变化率 1表示没有变化 0.8表示递减20%

	//对目标造成强制移动相关
	ForceMoveTime      float32 //强制移动时间
	ForceMoveSpeedSize float32 //强制移动速度大小
	ForceMoveLevel     int32   //强制移动等级
	ForceMoveBuff      string  //强制移动时的buff 随着移动结束消失
	ForceMoveType      int32   //强制移动类型

	//加血相关
	AddHPType  int32   //加血类型 0:不加 1:以AddHPValue为固定值 2:以AddHPValue为时间 加单位在此时间内受到的伤害值
	AddHPValue float32 //加血值

	PhysicalHurtAddHP float32 //物理伤害吸血 0.1表示 增加攻击造成伤害的10%的HP
	MagicHurtAddHP    float32 //魔法伤害吸血 0.1表示 增加攻击造成伤害的10%的HP

	PathHalo         string  //路径光环 在路径上创建光环
	PathHaloMinTime  float32 //路径光环的最短时间 1表示 相差1秒才创建光环
	PathHaloLastTime float64 //上次的路径光环时间

	//互换位置
	SwitchedPlaces     int32 //互换位置 1:是 2:否 只对目标为单位的情况生效
	DestForceAttackSrc int32 //目标强制攻击施法者 1:是 2:否

	//特殊情况处理 //3:风行束缚击
	Exception      int32  //0表示没有特殊情况
	ExceptionParam string //特殊情况处理参数

	IsSetStartPosition bool //是否设置了开始位置 如果设置了 在oncreate的时候就不设置了

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
	if src != nil {
		re.PhysicalHurtAddHP = src.PhysicalHurtAddHP
		re.MagicHurtAddHP = src.MagicHurtAddHP
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
	re.DestPos = vec2d.Vector3{pos.X, pos.Y, 1}
	if src != nil {
		re.PhysicalHurtAddHP = src.PhysicalHurtAddHP
		re.MagicHurtAddHP = src.MagicHurtAddHP
		re.DestPos = vec2d.Vector3{pos.X, pos.Y, src.ProjectileEndPosition.Y}
	}
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
	this.IsSetStartPosition = false

	this.NormalHurt.HurtType = 1 //物理伤害
	this.NormalHurt.HurtValue = 0

	this.OtherHurt = make([]HurtInfo, 0)
	this.Sputterings = make([]BulletSputtering, 0)
	this.TargetBuff = make([]BuffInfo, 0)
	this.TargetHalo = make([]HaloInfo, 0)
	this.HurtUnits = make(map[int32]*Unit)
	this.IsDoHurtOnMove = 2
	this.UnitTargetTeam = 2
	this.EveryDoHurtChangeHurtCR = 1

	this.Ejection = BulletEjection{0, 0, 1, make(map[int32]*Unit)}

	this.HurtRange.RangeType = 1 //单体攻击范围
	this.MoveType = 1            //瞬间移动
	this.UseUnitProjectilePos = 1

	this.Crit = 1
	this.ClearLevel = 0
	this.NoCareDodge = 0
	this.DoHurtPhysicalAmaorCV = 0

	this.AddHPType = 0
	this.AddHPValue = 0

	this.SetPathHalo("", 10)

	this.SwitchedPlaces = 2
	this.DestForceAttackSrc = 2

	this.Exception = 0
	this.ExceptionParam = ""

	this.SetForceMove(0, 0, 0, 0, "")
}

//设置弹射
func (this *Bullet) SetEjection(count int32, ejrange float32, ejdecay float32) {
	this.Ejection.EjectionCount = count
	this.Ejection.EjectionRange = ejrange
	this.Ejection.EjectionDecay = ejdecay
	if this.DestUnit != nil {
		this.Ejection.EjectionedTarget[this.DestUnit.ID] = this.DestUnit
	}

}

//设置路径光环
func (this *Bullet) SetPathHalo(ph string, mintime float32) {
	this.PathHalo = ph
	this.PathHaloMinTime = mintime
	this.PathHaloLastTime = 0
}

//创建路径光环
func (this *Bullet) DoCreatePathHalo() {
	if this.SrcUnit == nil || len(this.PathHalo) <= 0 {
		return
	}

	curtime := (utils.GetCurTimeOfSecond())
	//log.Info("-DoCreatePathHalo-:%f  :%f  :%f", curtime, this.PathHaloLastTime, this.PathHaloMinTime)
	if curtime-this.PathHaloLastTime > float64(this.PathHaloMinTime) {
		pos := vec2d.Vec2{X: this.Position.X, Y: this.Position.Y}

		this.SrcUnit.AddHaloFromStr(this.PathHalo, this.SkillLevel, &pos)

		this.PathHaloLastTime = curtime
		log.Info("DoCreatePathHalo succ :%f", curtime)

	}

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
func (this *Bullet) SetForceMove(time float32, speedsize float32, level int32, fmtype int32, buff string) {
	//对目标造成强制移动相关
	this.ForceMoveTime = time           //强制移动时间
	this.ForceMoveSpeedSize = speedsize //强制移动速度大小
	this.ForceMoveLevel = level         //强制移动等级
	this.ForceMoveBuff = buff
	this.ForceMoveType = fmtype

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

	if len(this.ModeType) <= 0 {
		this.ModeType = modetype
		this.Speed = speed
		if this.Speed <= 0 || this.Speed >= 1000000 {
			this.MoveType = 1
		} else {
			this.MoveType = 2
		}
	} else {
		if len(modetype) > 0 {
			this.ModeType += "," + modetype
		}
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

//设置startpos
func (this *Bullet) SetStartPosition(pos vec2d.Vector3) {
	this.Position = pos
	//开始位置
	this.StartPosition = this.Position.Clone()
	this.IsSetStartPosition = true
}

////计算初始位置
func (this *Bullet) OnCreate() {

	//起始相对位置为创建者的时候
	if this.StartPosOffsetObj == 1 {
		if this.SrcUnit == nil {
			this.Done()
			return
		}
		if this.IsSetStartPosition == false {
			if this.UseUnitProjectilePos == 1 {

				pos := this.SrcUnit.GetProjectileStartPos()

				dis1 := vec2d.Distanse(this.SrcUnit.Body.Position, vec2d.Vec2{X: pos.X, Y: pos.Y})
				dis2 := vec2d.Distanse(this.SrcUnit.Body.Position, vec2d.Vec2{X: this.DestPos.X, Y: this.DestPos.Y})
				if dis2 <= dis1 {
					this.Position = vec2d.Vector3{X: this.SrcUnit.Body.Position.X, Y: this.SrcUnit.Body.Position.Y, Z: pos.Z}
					//开始位置
					this.StartPosition = this.Position.Clone()
				} else {
					this.Position = pos
					//开始位置
					this.StartPosition = this.Position.Clone()
				}

			} else {

				this.Position = vec2d.NewVector3(this.SrcUnit.Body.Position.X, this.SrcUnit.Body.Position.Y, this.SrcUnit.ProjectileEndPosition.Y)
				//开始位置
				this.StartPosition = this.Position.Clone()
			}
		}

	}

	//如果为瞬间移动则直接计算伤害
	if this.MoveType == 1 {
		this.Position = this.DestPos
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
		var unit *Unit = nil
		if this.CallUnitTypeID > 0 {
			unit = CreateUnit(this.SrcUnit.InScene, this.CallUnitTypeID)

		} else if this.CallUnitTypeID == -1 {
			if this.SrcUnit != nil && this.SrcUnit.IsDisappear() == false {
				unit = CreateUnitByCopyUnit(this.SrcUnit, this.SrcUnit.MyPlayer)
			}

		}
		if unit == nil {
			continue
		}

		p1 := vec2d.Vec2{this.DestPos.X + float64(utils.GetRandomFloat(this.CallUnitOffsetPos)),
			this.DestPos.Y + float64(utils.GetRandomFloat(this.CallUnitOffsetPos))}

		log.Info("------------pos:%v", p1)
		unit.InitPosition = p1
		//
		unit.Camp = this.SrcUnit.Camp

		//buff
		unit.AddBuffFromStr(this.CallUnitBuff, this.CallUnitInfoSkillLevel, unit)
		unit.AddHaloFromStr(this.CallUnitHalo, this.CallUnitInfoSkillLevel, nil)

		unit.SetAI(NewNormalAI(unit))

		scene.NextAddUnit.Set(unit.ID, unit)
	}
}
func (this *Bullet) GetPosition2D() vec2d.Vec2 {
	return vec2d.Vec2{this.Position.X, this.Position.Y}
}

//处理溅射伤害
func (this *Bullet) DoSpurting(hurtvalue int32) {
	if this.DestUnit == nil || len(this.Sputterings) <= 0 {
		return
	}
	startpos := this.SrcUnit.Body.Position
	//瞬间到达以srcunit坐标为中心点 其他的以子弹当前点为中心点
	if this.MoveType == 1 {
		if this.SrcUnit == nil || this.SrcUnit.Body == nil {
			return
		}

	} else {
		startpos = vec2d.Vec2{this.Position.X, this.Position.Y}
	}

	allunit := this.SrcUnit.InScene.FindVisibleUnitsByPos(startpos)

	fangxiang := vec2d.Sub(vec2d.Vec2{this.Position.X, this.Position.Y}, startpos)
	if fangxiang.Length() <= 0 {
		fangxiang = vec2d.Sub(vec2d.Vec2{this.Position.X, this.Position.Y},
			vec2d.Vec2{this.StartPosition.X, this.StartPosition.Y})
		if fangxiang.Length() <= 0 {
			fangxiang = this.SrcUnit.Body.Direction
		}
	}
	for _, v := range allunit {
		//不能对目标造成溅射
		if v.IsDisappear() || v.Body == nil || v == this.DestUnit {
			continue
		}
		if this.SrcUnit.CheckIsEnemy(v) == false {
			continue
		}
		for _, v1 := range this.Sputterings {
			//技能免疫检测
			if v1.NoCareMagicImmune != 1 && v.MagicImmune == 1 {
				continue
			}
			t1 := fangxiang.GetNormalized()
			t1.MulToFloat64(float64(v1.Radius))
			if vec2d.CheckSector(startpos, t1, v1.Radian, v.Body.Position) == false {
				continue
			}

			//			dir := vec2d.Sub(v.Body.Position, startpos)
			//			if float32(dir.Length()) > v1.Radius {
			//				continue
			//			}
			//			if float32(vec2d.Angle(fangxiang, dir)) > v1.Radian {
			//				continue
			//			}
			log.Info("---DoSpurting--tt--:%d  :%f", hurtvalue, v1.HurtRatio)
			//造成溅射伤害
			this.SpurtingHurtUnit(v, int32(float32(hurtvalue)*v1.HurtRatio))
		}

	}
}

//溅射伤害 狂战斧 60%
func (this *Bullet) SpurtingHurtUnit(unit *Unit, value int32) {
	//log.Info("222222222")
	if unit == nil {
		return
	}
	//伤害 不会miss
	unit.BeAttackedFromValue(value, this.SrcUnit)

	//小于0 表示被miss 显示相关
	if this.SrcUnit == nil || this.SrcUnit.IsDisappear() {
		return
	}
	if this.SrcUnit.MyPlayer == nil {
		return
	}
	//为了显示 玩家造成的伤害
	mph := &protomsg.MsgPlayerHurt{HurtUnitID: unit.ID, HurtAllValue: value}
	this.SrcUnit.MyPlayer.AddHurtValue(mph)
}

func (this *Bullet) DoBuffException(buff *Buff) {
	if buff == nil || buff.Exception <= 0 {
		return
	}

	switch buff.Exception {
	case 4: //4:帕克大招
		{
			param := utils.GetFloat32FromString3(buff.ExceptionParam, ":")
			if len(param) < 3 {
				return
			}
			buff.SetConnectionPoint(vec2d.Vec2{X: this.Position.X, Y: this.Position.Y})
			//
		}
	default:
	}
}

//
func (this *Bullet) DoException(unit *Unit) {
	if this.Exception <= 0 || unit == nil || unit.IsDisappear() {
		return
	}
	//this.Exception = 0
	//this.ExceptionParam = ""

	switch this.Exception {
	case 3: //3:风行束缚击
		{
			//log.Info("1111111111111111111")
			if unit == nil || this.SrcUnit == nil || unit.IsDisappear() {
				return
			}
			//log.Info("222222222222222222:%s", this.ExceptionParam)
			param := utils.GetFloat32FromString3(this.ExceptionParam, ":")
			if len(param) < 5 {
				return
			}
			//log.Info("33333333333")
			little := param[0]
			big := param[1]
			distanse := param[2]
			buff := int(param[3])
			radian := float32(param[4])

			startpos := vec2d.Vec2{this.Position.X, this.Position.Y}
			allunit := this.SrcUnit.InScene.FindVisibleUnitsByPos(startpos)
			fangxiang := vec2d.Sub(startpos, vec2d.Vec2{this.StartPosition.X, this.StartPosition.Y})
			fangxiang.Normalize()
			fangxiang.MulToFloat64(float64(distanse))
			var nextunit *Unit = nil
			for _, v := range allunit {
				//不能对目标造成牵连
				if v.IsDisappear() || v.Body == nil || v == unit {
					continue
				}
				if this.SrcUnit.CheckIsEnemy(v) == false {
					continue
				}

				if vec2d.CheckSector(startpos, fangxiang, radian, v.Body.Position) == false {
					continue
				}
				nextunit = v
				break
			}
			if nextunit == nil {
				buffs := unit.AddBuffFromStr(strconv.Itoa(buff), this.SkillLevel, this.SrcUnit)
				if len(buffs) > 0 {
					buffs[0].RemainTime = little
					buffs[0].Time = little
				}
				//log.Info("4444444444")
			} else {
				buffs := unit.AddBuffFromStr(strconv.Itoa(buff), this.SkillLevel, this.SrcUnit)
				if len(buffs) > 0 {
					buffs[0].RemainTime = big
					buffs[0].Time = big
					buffs[0].SetConnectionPoint(nextunit.Body.Position)
				}

				buffs2 := nextunit.AddBuffFromStr(strconv.Itoa(buff), this.SkillLevel, this.SrcUnit)
				if len(buffs2) > 0 {
					buffs2[0].RemainTime = big
					buffs2[0].Time = big
					buffs2[0].SetConnectionPoint(unit.Body.Position)
				}
				//log.Info("5555555555")
			}

			//
		}
	case 6: //影魔影牙伤害叠加
		{
			if unit == nil || this.SrcUnit == nil || unit.IsDisappear() {
				return
			}
			//log.Info("222222222222222222:%s", this.ExceptionParam)
			param := utils.GetInt32FromString3(this.ExceptionParam, ":")
			if len(param) < 3 {
				return
			}
			//log.Info("33333333333")
			tagnum := int32(0)
			buff := unit.GetBuff(param[0])
			if buff != nil {
				tagnum = buff.TagNum
			}
			hurtvalue := tagnum * param[1]
			if hurtvalue > 0 {
				b := NewBullet1(this.SrcUnit, unit)
				b.SetProjectileMode("", 0)
				b.AddOtherHurt(HurtInfo{HurtType: param[2], HurtValue: int32(hurtvalue)})
				if b != nil {
					unit.AddBullet(b)
				}
			}

		}
	case 7: //斧王淘汰
		{
			if unit == nil || this.SrcUnit == nil || unit.IsDisappear() {
				return
			}
			//log.Info("222222222222222222:%s", this.ExceptionParam)
			param := utils.GetFloat32FromString3(this.ExceptionParam, ":")
			if len(param) < 3 {
				return
			}
			buffstr := strconv.Itoa(int(param[0]))
			hp := int32(param[1])
			castrange := param[2]
			//秒杀
			if unit.HP <= hp {
				unit.BeAttackedFromValue(-(hp + 1), this.SrcUnit)
				//buff
				allunit := this.SrcUnit.InScene.FindVisibleUnits(this.SrcUnit)
				for _, v := range allunit {
					if v == nil || v.Body == nil || v.IsDisappear() {
						continue
					}
					if v.UnitType != 1 || this.SrcUnit.CheckIsEnemy(v) == true {
						continue
					}
					dis := float32(vec2d.Distanse(this.SrcUnit.Body.Position, v.Body.Position))
					//log.Info("-----------------dis:%f", dis)
					if dis <= castrange {
						v.AddBuffFromStr(buffstr, this.SkillLevel, this.SrcUnit)
					}
				}
				//技能能却
				skilldata, ok := this.SrcUnit.Skills[this.SkillID]
				if ok {
					skilldata.RemainCDTime = 0
				}
				//
				this.OtherHurt = make([]HurtInfo, 0)
			}

		}
	default:
	}
}

//对单位造成伤害 只计算一次 hurtratio 伤害系数 狂战斧 60%
func (this *Bullet) HurtUnit(unit *Unit) int32 {
	//log.Info("222222222")
	if unit == nil {
		return 0
	}
	if _, ok := this.HurtUnits[unit.ID]; ok {
		return 0
	}
	//log.Info("33333333")
	this.DoException(unit)

	this.HurtUnits[unit.ID] = unit

	//对目标加血 目标为敌人则不能加血
	if this.AddHPType != 0 {
		if this.SrcUnit != nil {
			if this.SrcUnit.CheckIsEnemy(unit) == false {
				unit.DoAddHP(this.AddHPType, this.AddHPValue)
			}
		}

	}

	//伤害
	ismiss, hurtvalue, physichurt, magichurt := unit.BeAttacked(this)

	//强制移动
	if this.ForceMoveTime > 0 {

		if this.ForceMoveType == 1 || this.ForceMoveType == 4 {
			connectpoint := vec2d.Vec2{0, 0}
			if this.ForceMoveType == 1 { //向后推
				if this.HurtRange.RangeType == 1 {

					dir := vec2d.Sub(unit.Body.Position, vec2d.Vec2{this.StartPosition.X, this.StartPosition.Y})
					dir.Normalize()
					dir.MulToFloat64(float64(this.ForceMoveSpeedSize))
					unit.SetForceMove(this.ForceMoveTime, dir, this.ForceMoveLevel, float32(0))
					connectpoint = vec2d.Vec2{this.StartPosition.X, this.StartPosition.Y}
				} else {
					dir := vec2d.Sub(unit.Body.Position, this.GetPosition2D())
					dir.Normalize()
					dir.MulToFloat64(float64(this.ForceMoveSpeedSize))
					unit.SetForceMove(this.ForceMoveTime, dir, this.ForceMoveLevel, float32(0))
					connectpoint = this.GetPosition2D()
				}
			} else if this.ForceMoveType == 4 { //向前拉

				if this.HurtRange.RangeType == 1 {

					dir := vec2d.Sub(vec2d.Vec2{this.StartPosition.X, this.StartPosition.Y}, unit.Body.Position)
					dir.Normalize()
					dir.MulToFloat64(float64(this.ForceMoveSpeedSize))
					unit.SetForceMove(this.ForceMoveTime, dir, this.ForceMoveLevel, float32(0))
					connectpoint = vec2d.Vec2{this.StartPosition.X, this.StartPosition.Y}
				} else {
					dir := vec2d.Sub(this.GetPosition2D(), unit.Body.Position)
					dir.Normalize()
					dir.MulToFloat64(float64(this.ForceMoveSpeedSize))
					unit.SetForceMove(this.ForceMoveTime, dir, this.ForceMoveLevel, float32(0))
					connectpoint = this.GetPosition2D()
				}
			}

			//更改buff时间
			if len(this.ForceMoveBuff) > 0 {
				buffs := unit.AddBuffFromStr(this.ForceMoveBuff, this.SkillLevel, this.SrcUnit)
				for _, v := range buffs {
					v.RemainTime = this.ForceMoveTime
					v.Time = this.ForceMoveTime
					v.SetConnectionPoint(connectpoint)
				}
			}
		}

	}

	//小于0 表示被miss
	if ismiss == false {

		//驱散buff
		unit.ClearBuffForTarget(this.SrcUnit, this.ClearLevel)
		//buff
		for _, v := range this.TargetBuff {
			buffs := unit.AddBuffFromStr(v.Buff, v.BuffLevel, this.SrcUnit)
			//DoBuffException
			for _, v1 := range buffs {
				this.DoBuffException(v1)

				//目标强制攻击施法者
				if this.DestForceAttackSrc == 1 {
					//创建攻击命令
					v1.AttackTarget = this.SrcUnit
				}
			}
		}
		if this.SrcUnit == nil || this.SrcUnit.IsDisappear() {
			return hurtvalue
		}
		//吸血
		addhp := float32(0)
		pahp := this.PhysicalHurtAddHP + this.SrcUnit.PhysicalHurtAddHP
		if pahp > 0 {
			addhp += pahp * float32(-physichurt)
		}
		mahp := this.MagicHurtAddHP + this.SrcUnit.MagicHurtAddHP
		if mahp > 0 {
			addhp += mahp * float32(-magichurt)
		}
		if addhp > 0 {
			this.SrcUnit.ChangeHP(int32(addhp))
		}

		if this.SrcUnit.MyPlayer == nil {
			return hurtvalue
		}
		//为了显示 玩家造成的伤害
		mph := &protomsg.MsgPlayerHurt{HurtUnitID: unit.ID, HurtAllValue: hurtvalue}
		if this.Crit > 1 {
			mph.IsCrit = 1
		}
		this.SrcUnit.MyPlayer.AddHurtValue(mph)
	}
	return hurtvalue

}
func (this *Bullet) DoHalo() {
	//光环
	if len(this.TargetHalo) > 0 {
		if this.DestUnit != nil {
			if this.DestUnit.IsDisappear() == false {
				for _, v := range this.TargetHalo {
					halos := this.DestUnit.AddHaloFromStr(v.Halo, v.HaloLevel, &vec2d.Vec2{this.DestPos.X, this.DestPos.Y})
					for _, v1 := range halos {
						v1.CastUnit = this.SrcUnit
					}
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
		hurtvalue := this.HurtUnit(this.DestUnit)
		this.DoSpurting(hurtvalue)
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
			//UnitTargetTeam      int32   //目标单位关系 1:友方  2:敌方 3:友方敌方都行 5:自己 10:除自己外的其他 20 自己控制的单位(不包括自己)
			//			if this.UnitTargetTeam == 1 && this.SrcUnit.CheckIsEnemy(v) == true {
			//				continue
			//			}
			//			if this.UnitTargetTeam == 2 && this.SrcUnit.CheckIsEnemy(v) == false {
			//				continue
			//			}

			if this.SrcUnit.CheckUnitTargetTeam(v, this.UnitTargetTeam) == false {
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
func (this *Bullet) AddSputtering(sp BulletSputtering) {
	this.Sputterings = append(this.Sputterings, sp)
}

//伤害类型 (1:物理伤害 2:魔法伤害 3:纯粹伤害)
//计算攻击力  根据攻击类型
func (this *Bullet) GetAttackOfType(hurttype int32) int32 {

	val := int32(0)
	if this.NormalHurt.HurtType == hurttype {
		val += this.NormalHurt.HurtValue
		//每造成一次伤害 伤害递减
		this.NormalHurt.HurtValue = int32(float32(this.NormalHurt.HurtValue) * this.EveryDoHurtChangeHurtCR)
	}
	for k, v := range this.OtherHurt {
		if v.HurtType == hurttype {
			val += v.HurtValue
			//每造成一次伤害 伤害递减
			this.OtherHurt[k].HurtValue = int32(float32(v.HurtValue) * this.EveryDoHurtChangeHurtCR)
		}
	}
	//计算暴击
	if hurttype == 1 && this.Crit > 1 {
		val = int32(float32(val) * this.Crit)
	}

	return val
}

//处理交换位置
func (this *Bullet) DoSwitchedPlaces() {
	if this.SwitchedPlaces == 2 || this.DestUnit == nil || this.SrcUnit == nil || this.DestUnit.Body == nil || this.SrcUnit.Body == nil {
		return
	}
	srcpos := this.SrcUnit.Body.Position
	this.SrcUnit.Body.Position = this.DestUnit.Body.Position
	this.DestUnit.Body.Position = srcpos
}

//处理弹射
func (this *Bullet) DoEjection() {
	if this.Ejection.EjectionCount <= 0 || this.Ejection.EjectionRange <= 0 {
		return
	}
	if this.SrcUnit == nil || this.DestUnit == nil || this.SrcUnit.IsDisappear() {
		return
	}

	pos2d := vec2d.Vec2{X: this.Position.X, Y: this.Position.Y}
	allunit := this.SrcUnit.InScene.FindVisibleUnitsByPos(pos2d)
	for _, v := range allunit {
		if v.IsDisappear() {
			continue
		}
		//检测是否在范围内
		if v.Body == nil {
			continue
		}
		//目前只对敌人弹射
		if this.SrcUnit.CheckUnitTargetTeam(v, 2) == false {
			continue
		}
		//已经弹射过了
		_, ok := this.Ejection.EjectionedTarget[v.ID]
		if ok {
			continue
		}

		dis := float32(vec2d.Distanse(pos2d, v.Body.Position))
		if dis <= this.Ejection.EjectionRange {
			//可以弹射
			b := NewBullet1(this.SrcUnit, v)
			//设置坐标

			b.SetStartPosition(this.Position)
			//普通攻击衰减
			b.SetNormalHurtRatio(0)
			b.NormalHurt.HurtValue = int32(float32(this.NormalHurt.HurtValue) * this.Ejection.EjectionDecay)
			//额外伤害衰减
			for _, otherhurt := range this.OtherHurt {
				b.AddOtherHurt(HurtInfo{otherhurt.HurtType, int32(float32(otherhurt.HurtValue) * this.Ejection.EjectionDecay)})
			}
			//重新设置弹射信息
			//b.Ejection.EjectionedTarget = this.Ejection.EjectionedTarget
			b.SetEjection(this.Ejection.EjectionCount-1, this.Ejection.EjectionRange, this.Ejection.EjectionDecay)

			//设置弹道
			b.SetProjectileMode(this.ModeType, this.Speed)

			this.SrcUnit.AddBullet(b)
			return
		}
	}
}

//计算结果
func (this *Bullet) CalResult() {

	//处理召唤
	this.DoCallUnit()
	this.DoSwitchedPlaces()
	//计算伤害
	this.DoHurt()
	//处理弹射
	this.DoEjection()

	//等待删除
	this.Done()
}

//(1:创建,2:移动,3:到达后计算伤害 4:完成 可以删除了)
func (this *Bullet) Update(dt float32) {

	this.State = this.NextState

	if this.State == 1 {
		this.OnCreate()
		this.DoCreatePathHalo()
	} else if this.State == 2 {
		this.DoMove(dt)
		if this.IsDoHurtOnMove == 1 {
			this.DoHurt()
		}
		this.DoCreatePathHalo()
	} else if this.State == 3 {
		this.CalResult()
		this.DoCreatePathHalo()
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
