package gamecore

import (
	"dq/conf"
	"dq/cyward"
	"dq/db"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
	"dq/vec2d"
	"math"
	"strings"
)

var UnitID int32 = 1000000

func (this *Unit) TestData() {
	this.AttackRange = 2
	this.AttackRangeBuffer = 5

	this.Camp = 1
	this.IsDeath = 2

	this.AutoAttackTraceRange = 5
	this.AutoAttackTraceOutRange = 10

	this.HP = 1000
	this.MP = 1000
	this.MAX_HP = 1000
	this.MAX_MP = 1000
}

//获取1次攻击所需的时间 (141点攻击速度等于 1.20秒一次) 170/攻击速度 等于攻击时间
func (this *Unit) GetOneAttackTime() float32 {

	return 170.0 / this.AttackSpeed
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

//单位是否消失 (单位离线 单位死亡 单位在另一个空间:黑鸟的关..  开雾的状态下)
func (this *Unit) IsDisappear() bool {

	if this.IsDelete == true || this.IsDeath == 1 {
		return true
	}

	return false
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
	if at == nil || at.IsDisappear() {
		return
	}
	//判断阵营 攻击模式 是否允许本次攻击
	if this.AttackEnableForCampAndMode(at) == true {
		this.AttackCmdData = data
		this.AttackCmdDataTarget = at
		//log.Info("---------AttackCmd")
	}

}

//检查攻击指令的有效性 如果目标单位被场景删除 则无效
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
	///at := this.InScene.FindUnitByID(this.AttackCmdDataTarget.ID)
	if this.AttackCmdDataTarget.IsDisappear() {
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

//创建子弹的时候需要使用
type UnitProjectilePos struct {
	//弹道起始位置距离
	ProjectileStartPosDis float32
	//弹道起始位置高度
	ProjectileStartPosHeight float32
	//弹道结束位置距离
	ProjectileEndPosDis float32
	//弹道结束位置高度
	ProjectileEndPosHeight float32
}

//获取弹道起始位置
func (this *Unit) GetProjectileStartPos() vec2d.Vector3 {
	if this.Body == nil {
		return vec2d.NewVector3(0, 0, 0.5)
	}
	pos := vec2d.Add(this.Body.Position, vec2d.Mul(this.Body.Direction.GetNormalized(), float64(this.ProjectileStartPosDis)))

	//后期可能需要单位的z坐标参与计算
	return vec2d.NewVector3(pos.X, pos.Y, float64(this.ProjectileStartPosHeight))
}

//获取弹道结束位置
func (this *Unit) GetProjectileEndPos() vec2d.Vector3 {
	if this.Body == nil {
		return vec2d.NewVector3(0, 0, 0.5)
	}
	pos := vec2d.Add(this.Body.Position, vec2d.Mul(this.Body.Direction.GetNormalized(), float64(this.ProjectileEndPosDis)))

	//后期可能需要单位的z坐标参与计算
	return vec2d.NewVector3(pos.X, pos.Y, float64(this.ProjectileEndPosHeight))
}

//------------------单位本体------------------
type UnitProperty struct {
	conf.UnitFileData //单位配置文件数据
	UnitProjectilePos

	// 当前数据
	ControlID int32 //控制者ID
	IsMain    int32 //是否是主单位 1:是  2:不是

	AnimotorState int32 //动画状态 1:idle 2:walk 3:attack 4:skill 5:death
	//-------------新加----------
	AttackMode int32 //攻击模式(1:和平模式 2:组队模式 3:全体模式 4:阵营模式(玩家,NPC) 5:行会模式)

	IsDeath int32 //是否死亡(1:死亡 2:没死)
	Name    string

	//复合数据 会随时变动的数据 比如受buff影响攻击力降低  (每帧动态计算)
	HP            int32
	MAX_HP        int32
	MP            int32
	MAX_MP        int32
	Level         int32 //等级 会影响属性
	Experience    int32
	MaxExperience int32
	//-
	AttributeStrength     float32 //当前力量属性
	AttributeIntelligence float32 //当前智力属性
	AttributeAgility      float32 //当前敏捷属性
	//------攻击---------
	AttackSpeed float32 //攻击速度
	Attack      int32   //攻击力 (基础攻击力+属性影响+buff影响)
	AttackRange float32 //攻击范围 攻击距离
	MoveSpeed   float64 //移动速度
	MagicScale  float32 //技能增强
	MPRegain    float32 //魔法恢复
	//------防御---------
	PhysicalAmaor  float32 //物理护甲(-1)
	PhysicalResist float32 //物理伤害抵挡
	MagicAmaor     float32 //魔法抗性(0.25)
	StatusAmaor    float32 //状态抗性(0)
	Dodge          float32 //闪避(0)
	HPRegain       float32 //生命恢复

	//	IsDizzy     int8 //是否眩晕(1:眩晕 2:不眩晕)
	//	IsTwine     int8 //是否缠绕(1:缠绕 2:不缠绕)
	//	IsForceMove int8 //是否强制移动(1:强制移动 2:不强制移动) (推推棒等等)

}
type Unit struct {
	UnitProperty
	UnitCmd
	InScene  *Scene
	MyPlayer *Player
	ID       int32        //单位唯一ID
	Body     *cyward.Body //移动相关(位置信息) 需要设置移动速度
	State    UnitState    //逻辑状态
	AI       UnitAI

	IsDelete bool //是否被删除

	//发送数据部分
	ClientData    *protomsg.UnitDatas //客户端显示数据
	ClientDataSub *protomsg.UnitDatas //客户端显示差异数据
}

func (this *Unit) SetAI(ai UnitAI) {
	this.AI = ai

}

func CreateUnit(scene *Scene, typeid int32) *Unit {
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	//	文件数据
	unitre.UnitFileData = *(conf.GetUnitFileData(typeid))
	unitre.Name = unitre.UnitName
	unitre.Level = 1

	unitre.Init()
	unitre.IsMain = 2
	//unitre.UnitType = 2 //单位类型(1:英雄 2:普通单位 3:远古 4:boss)
	unitre.ControlID = -1

	return unitre
}

func CreateUnitByPlayer(scene *Scene, player *Player, datas []byte) *Unit {
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	unitre.MyPlayer = player

	//---------db.DB_CharacterInfo
	characterinfo := db.DB_CharacterInfo{}
	utils.Bytes2Struct(datas, &characterinfo)

	log.Info("---DB_CharacterInfo---%v", characterinfo)

	//名字 等级 经验
	unitre.Name = characterinfo.Name
	unitre.Level = characterinfo.Level
	unitre.Experience = characterinfo.Experience
	//	文件数据
	unitre.UnitFileData = *(conf.GetUnitFileData(characterinfo.Typeid))

	unitre.Init()
	unitre.IsMain = 1
	//unitre.UnitType = 1 //单位类型(1:英雄 2:普通单位 3:远古 4:boss)
	unitre.ControlID = player.Uid

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

	this.IsDeath = 2

	//满血 满蓝
	this.MAX_HP = this.BaseHP
	this.HP = this.MAX_HP
	this.MAX_MP = this.BaseMP
	this.MP = this.MAX_MP

	//弹道位置计算

	utils.GetFloat32FromString(this.ProjectileStartPos, &this.ProjectileStartPosDis, &this.ProjectileStartPosHeight)
	utils.GetFloat32FromString(this.ProjectileEndPos, &this.ProjectileEndPosDis, &this.ProjectileEndPosHeight)

	//this.TestData()
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

//AttributeStrength float32//当前力量属性
//	AttributeIntelligence float32//当前智力属性
//	AttributeAgility float32//当前敏捷属性
//计算属性(力量 智力 敏捷)
func (this *Unit) CalAttribute() {
	//基础力量+等级带来的力量成长
	this.AttributeStrength = this.AttributeBaseStrength + float32(this.Level-1)*this.AttributeStrengthGain

	this.AttributeIntelligence = this.AttributeBaseIntelligence + float32(this.Level-1)*this.AttributeIntelligenceGain

	this.AttributeAgility = this.AttributeBaseAgility + float32(this.Level-1)*this.AttributeAgilityGain

	//装备
	//技能
	//buff
}

//改变血量
func (this *Unit) ChangeHP(hp int32) {
	this.HP += hp
	if this.HP <= 0 {
		//死亡处理
		this.IsDeath = 1
	}
	if this.HP >= this.MAX_HP {
		this.HP = this.MAX_HP
	}
}

//改变MP
func (this *Unit) ChangeMP(mp int32) {
	this.MP += mp
	if this.MP <= 0 {
		this.MP = 0
	}
	if this.MP >= this.MAX_MP {
		this.MP = this.MAX_MP
	}
}

//计算MAX_HP和MAX_MP
func (this *Unit) CalMaxHP_MaxHP() {
	maxhp := this.BaseHP + int32(this.AttributeStrength*conf.StrengthAddHP)
	//装备
	//技能
	//buff

	if maxhp != this.MAX_HP {

		//按百分比增减当前血量
		changehp := float32(maxhp)/float32(this.MAX_HP)*float32(this.HP) - float32(this.HP)
		//log.Info("change hp:%d-----%d---%d----%d", int32(changehp), maxhp, this.MAX_HP, this.HP)
		this.MAX_HP = maxhp
		this.ChangeHP(int32(changehp))

	}

	//MP
	maxmp := this.BaseMP + int32(this.AttributeIntelligence*conf.IntelligenceAddMP)
	if maxmp != this.MAX_MP {

		changemp := float32(maxmp)/float32(this.MAX_MP)*float32(this.MP) - float32(this.MP)
		this.MAX_MP = maxmp
		this.ChangeMP(int32(changemp))

	}
}

//计算攻击速度
func (this *Unit) CalAttackSpeed() {
	//基础攻击加上敏捷增加的攻击
	this.AttackSpeed = float32(this.BaseAttackSpeed) + float32(this.AttributeAgility*conf.AgilityAddAttackSpeed)
	//装备
	//技能
	//buff

	//攻击速度取值范围
	if this.AttackSpeed <= 10 {
		this.AttackSpeed = 10
	} else if this.AttackSpeed >= float32(this.BaseMaxAttackSpeed) {
		this.AttackSpeed = float32(this.BaseMaxAttackSpeed)
	}

}

//计算攻击力
func (this *Unit) CalAttack() {
	//主属性(1:力量 2:敏捷 3:智力)
	//基础攻击力+主属性增减攻击力
	switch this.AttributePrimary {
	case 1:
		this.Attack = this.BaseAttack + int32(this.AttributeStrength*conf.AttributePrimaryAddAttack)
		break
	case 2:
		this.Attack = this.BaseAttack + int32(this.AttributeAgility*conf.AttributePrimaryAddAttack)
		break
	case 3:
		this.Attack = this.BaseAttack + int32(this.AttributeIntelligence*conf.AttributePrimaryAddAttack)
		break
	}

	//装备
	//技能
	//buff

}

//计算攻击距离
func (this *Unit) CalAttackRange() {
	//攻击范围
	this.AttackRange = this.BaseAttackRange

	//装备
	//技能
	//buff
}

//计算移动速度
func (this *Unit) CalMoveSpeed() {
	//基础移动速度+敏捷对移动速度的提升
	agilityaddspeed := float64(this.BaseMoveSpeed) * float64(this.AttributeAgility*conf.AgilityAddMoveSpeed)
	this.MoveSpeed = float64(this.BaseMoveSpeed) + agilityaddspeed

	//装备
	//技能
	//buff

	//移动行为逻辑速度设置
	this.Body.SpeedSize = this.MoveSpeed
}

//计算技能增强IntelligenceAddMagicScale
func (this *Unit) CalMagicScale() {
	//通过智力计算
	this.MagicScale = float32(this.AttributeIntelligence * conf.IntelligenceAddMagicScale)
	//装备
	//技能
	//buff
}

//计算魔法回复
func (this *Unit) CalMPRegain() {
	this.MPRegain = this.BaseMPRegain + float32(this.AttributeIntelligence*conf.IntelligenceAddMPRegain)
	//装备
	//技能
	//buff
}

////------防御---------
//	PhysicalAmaor  float32 //物理护甲(-1)
//	PhysicalResist float32 //物理伤害抵挡
//	MagicAmaor     float32 //魔法抗性(0.25)
//	StatusAmaor    float32 //状态抗性(0)
//	Dodge          float32 //闪避(0)
//	HPRegain       float32 //生命恢复
//计算护甲和物理抵抗
func (this *Unit) CalPhysicalAmaor() {
	//基础护甲+敏捷增减的护甲
	this.PhysicalAmaor = this.BasePhysicalAmaor + float32(this.AttributeAgility*conf.AgilityAddPhysicalAmaor)

	//装备
	//技能
	//buff

	//计算物理伤害抵挡
	this.PhysicalResist = 0.052 * this.PhysicalAmaor / (0.9 + 0.048*this.PhysicalAmaor)

}

//计算魔抗
func (this *Unit) CalMagicAmaor() {
	//非线性叠加
	//基础魔抗叠加力量带来的魔抗

	strenth := float32(this.AttributeStrength * conf.StrengthAddMagicAmaor)
	magicamaor := (1 - this.BaseMagicAmaor) * (1 - strenth)

	//装备
	//技能
	//buff

	this.MagicAmaor = 1 - magicamaor
}

//计算状态抗性
func (this *Unit) CalStatusAmaor() {
	//非线性叠加
	statusamaor := (1 - this.BaseStatusAmaor)

	//装备
	//技能
	//buff

	this.StatusAmaor = 1 - statusamaor
}

//计算闪避
func (this *Unit) CalDodge() {
	//非线性叠加
	dodge := (1 - this.BaseDodge)

	//装备
	//技能
	//buff

	this.Dodge = (1 - dodge)
}

//计算生命回复
func (this *Unit) CalHPRegain() {
	this.HPRegain = this.BaseHPRegain + float32(this.AttributeStrength*conf.StrengthAddHPRegain)
	//装备
	//技能
	//buff
}

//计算属性 (每一帧 都可能会变)
func (this *Unit) CalProperty() {
	//计算属性
	this.CalAttribute()
	//计算MAXHP MP
	this.CalMaxHP_MaxHP()
	//计算攻击速度
	this.CalAttackSpeed()
	//计算攻击力
	this.CalAttack()
	//计算攻击距离
	this.CalAttackRange()
	//计算移动速度
	this.CalMoveSpeed()
	//计算技能增强
	this.CalMagicScale()
	//计算魔法回复
	this.CalMPRegain()
	//计算护甲 物理伤害抵挡
	this.CalPhysicalAmaor()
	//计算魔抗
	this.CalMagicAmaor()
	//计算状态抗性
	this.CalStatusAmaor()
	//计算闪避
	this.CalDodge()
	//计算生命回复
	this.CalHPRegain()

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

	this.ClientData.UnitType = this.UnitType
	this.ClientData.AttackAcpabilities = this.AttackAcpabilities
	this.ClientData.AttackMode = this.AttackMode
	this.ClientData.IsMain = this.IsMain

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

	this.ClientDataSub.UnitType = this.UnitType - this.ClientData.UnitType
	this.ClientDataSub.AttackAcpabilities = this.AttackAcpabilities - this.ClientData.AttackAcpabilities
	this.ClientDataSub.AttackMode = this.AttackMode - this.ClientData.AttackMode

	this.ClientDataSub.IsMain = this.IsMain - this.ClientData.IsMain

}

//被删除的时候
func (this *Unit) OnDestroy() {
	this.IsDelete = true
}

//即时属性获取
