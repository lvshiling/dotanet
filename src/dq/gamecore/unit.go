package gamecore

import (
	"dq/cyward"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"math"
	"strings"
)

var UnitID int32 = 1000000

//单位配置文件数据
type UnitFileData struct {
	//配置文件数据
	TypeID                    int32   //类型ID
	ModeType                  string  //模型
	BaseHP                    int32   //基础HP
	BaseMP                    int32   //基础MP
	BaseAttackSpeed           int32   //基础攻击速度(141点攻击速度等于 1.20秒一次)
	BaseAttack                int32   //基础攻击力
	BaseAttackRange           float32 //基础攻击范围
	BaseMoveSpeed             float32 //基础移动速度
	BaseMagicScale            float32 //基础技能增强
	BaseMPRegain              float32 //基础魔法恢复
	BasePhysicalAmaor         float32 //基础物理护甲(-1)
	BaseMagicAmaor            float32 //基础魔法抗性(0.25)
	BaseStatusAmaor           float32 //基础状态抗性(0)
	BaseDodge                 float32 //基础闪避(0)
	BaseHPRegain              float32 //基础生命恢复
	AttributePrimary          int8    //主属性(1:力量 2:敏捷 3:智力)
	AttributeBaseStrength     float32 //基础力量
	AttributeStrengthGain     float32 //力量成长
	AttributeBaseIntelligence float32 //基础智力
	AttributeIntelligenceGain float32 //智力成长
	AttributeBaseAgility      float32 //基础敏捷
	AttributeAgilityGain      float32 //敏捷成长
	AttackAnimotionPoint      float32 //攻击前摇(0.3)
	AttackRangeBuffer         float32 //前摇不中断攻击范围
	ProjectileMode            string  //弹道模型
	ProjectileSpeed           float32 //弹道速度
	UnitType                  int8    //单位类型(1:英雄 2:普通单位 3:远古 4:boss)
	AttackAcpabilities        int8    //(1:近程攻击 2:远程攻击)

	//-------------新加----
	AutoAttackTraceRange    float32 //自动攻击的追击范围
	AutoAttackTraceOutRange float32 //自动攻击的取消追击范围
	//-----
	Camp int8 //阵营(1:玩家 2:NPC)
}

func (this *Unit) TestData() {
	this.AttackRange = 2
	this.AttackRangeBuffer = 5

	this.Camp = 1
	this.IsDeath = 2

	this.AutoAttackTraceRange = 5
	this.AutoAttackTraceOutRange = 10
}

type UnitProperty struct {
	UnitFileData //单位配置文件数据

	//基础数据 当前数据
	HP            int32
	MAX_HP        int32
	MP            int32
	MAX_MP        int32
	Name          string
	Level         int32
	Experience    int32
	MaxExperience int32

	ControlID int32 //控制者ID

	AnimotorState int32 //动画状态 1:idle 2:walk 3:attack 4:skill 5:death

	//复合数据 会随时变动的数据 比如受buff影响攻击力降低  (每帧动态计算)
	MoveSpeed   float64
	Attack      int32   //攻击力 (基础攻击力+属性影响+buff影响)
	AttackRange float32 //攻击范围

	//-------------新加----------
	AttackMode int8 //攻击模式(1:和平模式 2:组队模式 3:全体模式 4:阵营模式(玩家,NPC) 5:行会模式)

	IsDeath int8 //是否死亡(1:死亡 2:没死)

	//	IsDizzy     int8 //是否眩晕(1:眩晕 2:不眩晕)
	//	IsTwine     int8 //是否缠绕(1:缠绕 2:不缠绕)
	//	IsForceMove int8 //是否强制移动(1:强制移动 2:不强制移动) (推推棒等等)

}

//获取1次攻击所需的时间 (141点攻击速度等于 1.20秒一次)
func (this *Unit) GetOneAttackTime() float32 {
	return 1.5
}

//目标离自动攻击范围的距离 小于0 表示在内
func (this *Unit) GetDistanseOfAutoAttackRange(target *Unit) float64 {
	if target == nil {
		return 10
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()

	return targetdis - float64(this.AutoAttackTraceRange)
}

//目标是否脱离 自动攻击取消追击范围
func (this *Unit) IsOutAutoAttackTraceOutRange(target *Unit) bool {

	if target == nil {
		return true
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()
	if targetdis <= float64(this.AutoAttackTraceOutRange) {
		return false
	}

	return true
}

//目标是否脱离 前摇中断范围
func (this *Unit) IsOutAttackRangeBuffer(target *Unit) bool {

	if target == nil {
		return true
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()
	if targetdis <= float64(this.AttackRange)+float64(this.AttackRangeBuffer) {
		return false
	}

	return true
}

//单位是否能被攻击 (各种BUF状态 无敌状态 免疫状态)
func (this *Unit) IsCanBeAttack() bool {
	return true
}

//目标是否在攻击范围内
func (this *Unit) IsInAttackRange(target *Unit) bool {
	//this.Body.Position.X
	if target == nil {
		return false
	}

	dir := vec2d.Sub(this.Body.Position, target.Body.Position)

	targetdis := dir.Length()
	if targetdis <= float64(this.AttackRange) {
		return true
	}

	return false
}

//设置动画状态
func (this *Unit) SetAnimotorState(anistate int32) {
	this.AnimotorState = anistate

}

//设置单位状态
func (this *Unit) SetState(state UnitState) {
	this.State = state

}

//命令操作相关
type UnitCmd struct {
	//移动命令
	MoveCmdData *protomsg.CS_PlayerMove
	//攻击命令
	AttackCmdData *protomsg.CS_PlayerAttack
	//攻击目标
	AttackCmdDataTarget *Unit
}

//是否有攻击命令
func (this *Unit) HaveAttackCmd() bool {

	if this.AttackCmdData != nil {
		return true
	}
	return false
}

//能否攻击(根据阵营,攻击模式判断与是否死亡)
func (this *Unit) AttackEnableForCampAndMode(target *Unit) bool {
	if this.Camp != target.Camp {
		return true
	}
	if this.IsDeath == 1 || target.IsDeath == 1 {
		return false
	}
	//攻击模式(1:和平模式 2:组队模式 3:全体模式 4:阵营模式(玩家,NPC) 5:行会模式)
	if this.AttackMode == 1 {
		return false
	} else if this.AttackMode == 3 {
		//全体模式下 不是同一个玩家控制的单位可以攻击
		if this.ControlID != target.ControlID {
			return true
		} else {
			return false
		}

	}

	return true

}

//攻击行为命令
func (this *Unit) AttackCmd(data *protomsg.CS_PlayerAttack) {

	if this.InScene == nil {
		return
	}
	at := this.InScene.FindUnitByID(data.TargetUnitID)
	if at == nil {
		return
	}
	//判断阵营 攻击模式 是否允许本次攻击
	if this.AttackEnableForCampAndMode(at) == true {
		this.AttackCmdData = data
		this.AttackCmdDataTarget = at
		//log.Info("---------AttackCmd")
	}

}

//检查攻击指令的有效性 如果目标单位被场景删除 则无需
func (this *Unit) CheckAttackCmd() {
	if this.InScene == nil {
		return
	}
	if this.HaveAttackCmd() == false {
		return
	}
	if this.AttackCmdDataTarget == nil {
		return
	}
	at := this.InScene.FindUnitByID(this.AttackCmdDataTarget.ID)
	if at == nil {
		this.StopAttackCmd()
	}
}

//中断攻击命令
func (this *Unit) StopAttackCmd() {
	this.AttackCmdData = nil
	this.AttackCmdDataTarget = nil
}

//是否有移动命令
func (this *Unit) HaveMoveCmd() bool {

	if this.MoveCmdData != nil && this.MoveCmdData.IsStart == true {
		return true
	}

	return false
}

//是否可以移动
func (this *Unit) GetCanMove() bool {

	//	if this.IsDizzy == 1 || this.IsTwine == 1 || this.IsForceMove == 1 {
	//		return false
	//	}

	return true
}

//设置单位朝向
func (this *Unit) SetDirection(dir vec2d.Vec2) {
	this.Body.Direction = dir
}

//移动行为命令
func (this *Unit) MoveCmd(data *protomsg.CS_PlayerMove) {

	if this.AttackCmdData == nil {
		this.MoveCmdData = data
		return
	}
	//检测是否要中断攻击
	if this.MoveCmdData == nil {
		this.StopAttackCmd()
		this.MoveCmdData = data
	} else {
		if this.MoveCmdData.IsStart == false && data.IsStart == true {
			this.StopAttackCmd()
			this.MoveCmdData = data
		}
		if this.MoveCmdData.IsStart == true && data.IsStart == true {
			v1 := vec2d.Vec2{X: float64(this.MoveCmdData.X), Y: float64(this.MoveCmdData.Y)}
			v2 := vec2d.Vec2{X: float64(data.X), Y: float64(data.Y)}

			angle := vec2d.Angle(v1, v2)

			if math.Abs(angle) >= 0.4 {
				this.StopAttackCmd()
				this.MoveCmdData = data
				log.Info("---------angle:%f", angle)
			}
		}
	}

	log.Info("---------MoveCmd")
}

//------------------单位本体------------------
type Unit struct {
	UnitProperty
	UnitCmd
	InScene  *Scene
	MyPlayer *Player
	ID       int32        //单位唯一ID
	Body     *cyward.Body //移动相关(位置信息) 需要设置移动速度
	State    UnitState    //逻辑状态
	AI       UnitAI

	//发送数据部分
	ClientData    *protomsg.UnitDatas //客户端显示数据
	ClientDataSub *protomsg.UnitDatas //客户端显示差异数据
}

func (this *Unit) SetAI(ai UnitAI) {
	this.AI = ai

}

func CreateUnit(scene *Scene) *Unit {
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	unitre.Init()

	return unitre
}

func CreateUnitByPlayer(scene *Scene, player *Player, datas []byte) *Unit {
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	unitre.MyPlayer = player
	utils.Bytes2Struct(datas, &unitre.UnitProperty)
	unitre.Init()

	return unitre
}

//初始化
func (this *Unit) Init() {
	this.State = NewIdleState(this)

	//设置一些初始状态
	//	this.IsDizzy = 2
	//	this.IsTwine = 2
	//	this.IsForceMove = 2
	this.AttackMode = 1 //和平攻击模式

	this.TestData()
}

//
//更新 范围影响的buff
func (this *Unit) PreUpdate(dt float64) {

}

//更新
func (this *Unit) Update(dt float64) {
	//设置是否有碰撞  设置移动速度 和逻辑状态

	//

	this.CalProperty()
	//AI
	if this.AI != nil {
		this.AI.Update(dt)
	}
	this.CheckAttackCmd()

	//逻辑状态更新
	this.State.OnTransform()
	this.State.Update(dt)
}

//计算属性 (每一帧 都可能会变)
func (this *Unit) CalProperty() {

	this.MoveSpeed = float64(this.BaseMoveSpeed)

	this.Body.SpeedSize = this.MoveSpeed
}

//刷新客户端显示数据
func (this *Unit) FreshClientData() {

	if this.ClientData == nil {
		this.ClientData = &protomsg.UnitDatas{}
	}

	this.ClientData.HP = this.HP
	this.ClientData.MaxHP = this.MAX_HP
	this.ClientData.MP = this.MP
	this.ClientData.MaxMP = this.MAX_MP
	this.ClientData.Name = this.Name
	this.ClientData.Level = this.Level
	this.ClientData.ID = this.ID
	this.ClientData.ModeType = this.ModeType
	this.ClientData.Experience = this.Experience
	this.ClientData.MaxExperience = this.MaxExperience
	this.ClientData.ControlID = this.ControlID
	this.ClientData.AnimotorState = this.AnimotorState
	this.ClientData.AttackTime = this.GetOneAttackTime()

	this.ClientData.X = float32(this.Body.Position.X)
	this.ClientData.Y = float32(this.Body.Position.Y)

	this.ClientData.DirectionX = float32(this.Body.Direction.X)
	this.ClientData.DirectionY = float32(this.Body.Direction.Y)
}

//func (this *Unit) OnePropSub(prop interface{}){

//}

//刷新客户端显示差异数据
func (this *Unit) FreshClientDataSub() {

	if this.ClientDataSub == nil {
		this.ClientDataSub = &protomsg.UnitDatas{}
	}
	if this.ClientData == nil {
		this.FreshClientData()
		*this.ClientDataSub = *this.ClientData
		return
	}

	//
	//字符串部分
	if strings.Compare(this.Name, this.ClientData.Name) != 0 {
		this.ClientDataSub.Name = this.Name
	} else {
		this.ClientDataSub.Name = ""
	}
	if strings.Compare(this.ModeType, this.ClientData.ModeType) != 0 {
		this.ClientDataSub.ModeType = this.ModeType
	} else {
		this.ClientDataSub.ModeType = ""
	}
	//
	if this.AnimotorState != this.ClientData.AnimotorState {
		this.ClientDataSub.AnimotorState = this.AnimotorState
	} else {
		this.ClientDataSub.AnimotorState = 0
	}
	//攻击
	if this.AnimotorState == 3 {
		this.ClientData.AttackTime = this.GetOneAttackTime()
	} else {
		this.ClientData.AttackTime = 0
	}

	//当前数据与上一次数据对比 相减 数值部分
	this.ClientDataSub.HP = this.HP - this.ClientData.HP
	this.ClientDataSub.MaxHP = this.MAX_HP - this.ClientData.MaxHP
	this.ClientDataSub.MP = this.MP - this.ClientData.MP
	this.ClientDataSub.MaxMP = this.MAX_MP - this.ClientData.MaxMP
	this.ClientDataSub.Level = this.Level - this.ClientData.Level
	this.ClientDataSub.Experience = this.Experience - this.ClientData.Experience
	this.ClientDataSub.MaxExperience = this.MaxExperience - this.ClientData.MaxExperience
	this.ClientDataSub.ControlID = this.ControlID - this.ClientData.ControlID
	this.ClientDataSub.X = float32(this.Body.Position.X) - this.ClientData.X
	this.ClientDataSub.Y = float32(this.Body.Position.Y) - this.ClientData.Y

	this.ClientDataSub.DirectionX = float32(this.Body.Direction.X) - this.ClientData.DirectionX
	this.ClientDataSub.DirectionY = float32(this.Body.Direction.Y) - this.ClientData.DirectionY
}

//即时属性获取
