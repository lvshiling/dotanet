package gamecore

import (
	"dq/vec2d"
)

var BulletID int32 = 100

func GetBulletID() int32 {
	BulletID++
	if BulletID >= 100000000 {
		BulletID = 100
	}
	return BulletID
}

type Bullet struct {
	ID       int32
	SrcUnit  *Unit         //施法单位
	DestUnit *Unit         //目标单位
	DestPos  vec2d.Vector3 //目的地位置

	Position vec2d.Vector3 //子弹的当前位置

	ModeType string  //子弹模型
	Speed    float32 //子弹速度

	State int32 //子弹状态(1:创建,2:移动,3:到达后计算伤害 4:完成 可以删除了)

	StartPosOffsetObj int32 //起始位置参照物   1创建者 2目标 一般情况都为1创建者  像从天而降的闪电就为2目标 的头顶

	SkillID    int32 //技能ID  如果技能ID为 -1表示普通攻击
	SkillLevel int32 //技能等级
}

func NewBullet1(src *Unit, dest *Unit) *Bullet {
	re := &Bullet{}
	re.SrcUnit = src
	re.DestUnit = dest
	re.Speed = 5
	re.State = 0
	re.SkillID = -1
	re.StartPosOffsetObj = 1
	//唯一ID处理
	re.ID = GetBulletID()

	return re
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

	}

	this.State = 1
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
		this.State = 2
		return
	} else {
		//t1 := vec2d.Sub3(this.DestPos, this.Position)
		movevec3 := vec2d.Normalize3(vec2d.Sub3(this.DestPos, this.Position))
		movevec3.Multiply(movedis)
		//movevec3 := vec2d.Normalize3(vec2d.Sub3(this.DestPos, this.Position)).Multiply(movedis)
		this.Position.Add(movevec3)
	}
}

//更新
func (this *Bullet) Update(dt float32) {
	if this.State == 0 {
		this.OnCreate()
	} else if this.State == 1 {
		this.DoMove(dt)
	}
}

//子弹done
func (this *Bullet) Done() {
	this.State = 4
}
