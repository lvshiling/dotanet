package gamecore

import (
	"dq/cyward"
	"dq/log"
	"dq/protobuf"
	"dq/utils"
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
}

type UnitProperty struct {
	UnitFileData

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
}

type UnitCmd struct {
	Move *protomsg.CS_PlayerMove
}

type Unit struct {
	UnitProperty
	UnitCmd
	InScene  *Scene
	MyPlayer *Player
	ID       int32        //单位唯一ID
	Body     *cyward.Body //移动相关(位置信息) 需要设置移动速度
	State    UnitState    //逻辑状态

	//发送数据部分
	ClientData    *protomsg.UnitDatas //客户端显示数据
	ClientDataSub *protomsg.UnitDatas //客户端显示差异数据
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
}

//
//更新 范围影响的buff
func (this *Unit) PreUpdate(dt float64) {

}

//移动行为命令
func (this *Unit) MoveCmd(data *protomsg.CS_PlayerMove) {
	this.Move = data

	log.Info("---------MoveCmd")
}

//更新
func (this *Unit) Update(dt float64) {
	//设置是否有碰撞  设置移动速度 和逻辑状态

	//

	this.CalProperty()
	//逻辑状态更新
	this.State.OnTransform()
	this.State.Update(dt)
}

func (this *Unit) CalProperty() {

	this.MoveSpeed = this.BaseMoveSpeed

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

	this.ClientData.X = float32(this.Body.Position.X)
	this.ClientData.Y = float32(this.Body.Position.Y)
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
}

//即时属性获取

//是否可以移动
func (this *Unit) GetCanMove() bool {
	return true
}
