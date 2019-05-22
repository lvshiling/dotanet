package gamecore

import (
	//"dq/log"
	"dq/protobuf"
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

//技能无视闪避  普工攻击要计算闪避
type Bullet struct {
	ID       int32
	SrcUnit  *Unit         //施法单位
	DestUnit *Unit         //目标单位
	DestPos  vec2d.Vector3 //目的地位置

	Position      vec2d.Vector3 //子弹的当前位置
	StartPosition vec2d.Vector3 //子弹的初始位置

	ModeType string  //子弹模型
	Speed    float32 //子弹速度
	MoveType int32   //移动类型 1:瞬间移动  2:直线移动

	State     int32 //子弹状态(1:创建,2:移动,3:到达后计算结果(伤害和回血) 4:完成 可以删除了)
	NextState int32 //下一帧状态

	StartPosOffsetObj int32 //起始位置参照物   1创建者 2目标 一般情况都为1创建者  像从天而降的闪电就为2目标 的头顶

	SkillID int32 //技能ID  如果技能ID为 -1表示普通攻击
	//SkillLevel int32 //技能等级

	NormalHurt HurtInfo   //攻击伤害(以英雄攻击力计算伤害) 值为计算暴击后的值
	OtherHurt  []HurtInfo //其他伤害也就是额外伤害

	HurtRange BulletRange //范围

	//--------附加攻击特效------

	//发送数据部分
	ClientData    *protomsg.BulletDatas //客户端显示数据
	ClientDataSub *protomsg.BulletDatas //客户端显示差异数据
}

func NewBullet1(src *Unit, dest *Unit) *Bullet {
	re := &Bullet{}
	re.SrcUnit = src
	re.DestUnit = dest

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

	this.NormalHurt.HurtType = 1 //物理伤害
	this.NormalHurt.HurtValue = 0

	this.OtherHurt = make([]HurtInfo, 0)

	this.HurtRange.RangeType = 1 //单体攻击范围
	this.MoveType = 1            //瞬间移动
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

		if this.DestUnit != nil {
			this.DestPos = this.DestUnit.GetProjectileEndPos()
		}

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
		if this.DestUnit.IsDisappear() == false {
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

//计算伤害
func (this *Bullet) DoHurt() {
	//获取到受伤害的单位 (狂战斧攻击特效也会影响单位数量 )

	if this.HurtRange.RangeType == 1 {
		//单体范围
		if this.DestUnit == nil {
			return
		}
		//伤害
		this.DestUnit.BeAttacked(this)

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

	return val
}

//计算结果
func (this *Bullet) CalResult() {

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
