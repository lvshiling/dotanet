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
	//死亡后不能攻击
	if this.IsDeath == 1 {
		return false
	}
	return true
}

//是否能看见目标 目标可能隐身 本单位可能能看见隐身单位
func (this *Unit) CanSeeTarget(target *Unit) bool {
	if target == nil {
		return false
	}
	//目标隐身
	if target.Invisible == 1 {
		//目标是敌人
		if this.CheckIsEnemy(target) == true {
			//log.Info("----no see")
			return false
		}
		//log.Info("----Invisible")
	}
	return true
}

//单位是否消失 (单位离线 单位死亡 )
func (this *Unit) IsDisappear() bool {

	if this.IsDelete == true || this.IsDeath == 1 || this.InScene == nil {
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
	//技能命令
	SkillCmdData *protomsg.CS_PlayerSkill
}

//-----------------------技能命令--------------------
//是否有技能命令
func (this *Unit) HaveSkillCmd() bool {

	if this.SkillCmdData != nil {
		return true
	}
	return false
}

//检查目标释放在技能施法范围内
func (this *Unit) IsInSkillRange(data *protomsg.CS_PlayerSkill) bool {
	//检查本单位是否有这个技能
	skilldata, ok := this.Skills[data.SkillID]
	if ok == false {
		return false
	}
	//技能施法目标类型 为 单位
	if skilldata.CastTargetType == 2 {
		target := this.InScene.FindUnitByID(data.TargetUnitID)

		dir := vec2d.Sub(this.Body.Position, target.Body.Position)

		targetdis := dir.Length()
		if targetdis > float64(skilldata.CastRange)+float64(this.AddedMagicRange) {
			return false
		}
	} else if skilldata.CastTargetType == 3 { //以目的点为施法目标
		targetpos := vec2d.Vec2{X: float64(data.X), Y: float64(data.Y)}

		dir := vec2d.Sub(this.Body.Position, targetpos)

		targetdis := dir.Length()
		if targetdis > float64(skilldata.CastRange)+float64(this.AddedMagicRange) {
			return false
		}
	}

	return true

}

//检查技能是否可以对目标释放
func (this *Unit) CheckCastSkillTarget(target *Unit, skilldata *Skill) bool {
	//int32 UnitTargetTeam = 8;//目标单位关系 1:友方  2:敌方 3:友方敌方都行
	//int32 UnitTargetCamp = 9;//目标单位阵营 (1:玩家 2:NPC) 3:玩家NPC都行
	if target == nil {
		return false
	}
	//目标消失
	if target.IsDisappear() == true || this.CanSeeTarget(target) == false {
		return false
	}

	isEnemy := this.CheckIsEnemy(target)
	//目标技能免疫 且目标是敌人
	if isEnemy == true {
		if target.MagicImmune == 1 {
			if skilldata.NoCareMagicImmune == 2 {
				return false
			}
		}
	}

	//与目标单位的关系
	if skilldata.UnitTargetTeam == 1 {
		if isEnemy == true {
			return false
		}
	} else if skilldata.UnitTargetTeam == 2 {
		if isEnemy == false {
			return false
		}
	}

	return true
	//if skilldata.UnitTargetCamp
}

//检查攻击 触发攻击特效
func (this *Unit) CheckTriggerAttackSkill(b *Bullet) {
	for _, v := range this.Skills {
		//CastType              int32   // 施法类型:  1:主动技能  2:被动技能
		//TriggerTime int32 //触发时间 0:表示不触发 1:攻击时 2:被攻击时
		//主动技能
		if v.CastType == 2 && v.TriggerTime == 1 {
			//检查cd 魔法消耗
			if v.RemainCDTime <= 0 {
				//检查 触发概率
				if utils.CheckRandom(v.TriggerProbability) {
					//触发
					//添加自己的buff
					this.AddBuffFromStr(v.MyBuff, v.Level, this)
					//暴击
					b.SetCrit(v.TriggerCrit)
					//召唤信息
					//召唤信息
					b.BulletCallUnitInfo = BulletCallUnitInfo{v.CallUnitInfo, v.Level}
					//目标buff
					b.AddTargetBuff(v.TargetBuff, v.Level)
					//强制移动
					b.SetForceMove(v.ForceMoveTime, v.ForceMoveSpeedSize, v.ForceMoveLevel)
				}
			}
		} else if v.CastType == 1 && v.CastTargetType == 4 && v.AttackAutoActive == 1 {
			//主动技能 攻击时自动释放的攻击特效
			if v.RemainCDTime > 0 {
				continue
			}
			if this.SkillEnable != 1 {
				continue
			}
			if b.DestUnit == nil {
				continue
			}
			if b.DestUnit.MagicImmune == 1 {
				if v.NoCareMagicImmune == 2 {
					continue
				}
			}
			//目标buff
			b.AddTargetBuff(v.TargetBuff, v.Level)
			//强制移动
			b.SetForceMove(v.ForceMoveTime, v.ForceMoveSpeedSize, v.ForceMoveLevel)

		}
	}
}

//使用技能 创建子弹
func (this *Unit) DoSkill(data *protomsg.CS_PlayerSkill, targetpos vec2d.Vec2) {

	//检查本单位是否有这个技能
	skilldata, ok := this.Skills[data.SkillID]
	if ok == false {
		return
	}

	//驱散自己的buff
	this.ClearBuffForTarget(this, skilldata.MyClearLevel)

	//MyBuff
	this.AddBuffFromStr(skilldata.MyBuff, skilldata.Level, this)
	//MyHalo
	this.AddHaloFromStr(skilldata.MyHalo, skilldata.Level, nil)

	//创建子弹
	b := skilldata.CreateBullet(this, data)
	if b != nil {
		b.ClearLevel = skilldata.TargetClearLevel //设置驱散等级
		if skilldata.TriggerAttackEffect == 1 {
			this.CheckTriggerAttackSkill(b)
		}
		this.AddBullet(b)
	}
	//BlinkToTarget
	if skilldata.BlinkToTarget == 1 {
		this.Body.BlinkToPos(targetpos, 0)
	}

	//消耗 CD
	namacost := skilldata.ManaCost - int32(this.ManaCost*float32(skilldata.ManaCost))
	this.ChangeMP(-namacost)

	cdtime := skilldata.Cooldown - this.MagicCD*skilldata.Cooldown
	skilldata.FreshCDTime(cdtime)

	//删除技能命令
	this.StopSkillCmd()

	//如果目标是敌人 则自动攻击
	targetunit := this.InScene.FindUnitByID(data.TargetUnitID)
	if targetunit != nil {
		if this.CheckIsEnemy(targetunit) == true {
			acd := &protomsg.CS_PlayerAttack{}
			acd.TargetUnitID = data.TargetUnitID
			this.AttackCmd(acd)
		}
	}

}

//检查是否能使用技能
func (this *Unit) UseSkillEnable(data *protomsg.CS_PlayerSkill) bool {
	if this.SkillEnable == 2 {
		return false
	}

	if this.IsDisappear() == true {
		return false
	}

	//检查本单位是否有这个技能
	skilldata, ok := this.Skills[data.SkillID]
	if ok == false {
		return false
	}
	//技能等级
	if skilldata.Level <= 0 {
		return false
	}

	//不是主动技能
	if skilldata.CastType != 1 {
		return false
	}
	//cd中
	if skilldata.RemainCDTime > 0 {
		return false
	}
	//魔法不足
	shouldmp := skilldata.ManaCost - int32(this.ManaCost*float32(skilldata.ManaCost))
	if shouldmp > this.MP {
		return false
	}

	//技能施法目标类型 为 单位
	if skilldata.CastTargetType == 2 {
		target := this.InScene.FindUnitByID(data.TargetUnitID)
		if this.CheckCastSkillTarget(target, skilldata) == false {
			return false
		}
	}
	//施法距离就不判断了 如果距离不够单位自己移动过去

	return true
}

//技能行为命令
func (this *Unit) SkillCmd(data *protomsg.CS_PlayerSkill) {

	//如果是攻击特效技能(比如小黑的冰箭)
	//检查本单位是否有这个技能
	skilldata, ok := this.Skills[data.SkillID]
	if ok == false {
		return
	}
	if skilldata.CastType == 1 && skilldata.CastTargetType == 4 {
		skilldata.DoActive()
		return
	}

	//判断阵营 攻击模式 是否允许本次攻击
	if this.UseSkillEnable(data) == true {
		this.SkillCmdData = data
		//this.AttackCmdDataTarget = at
		log.Info("---------SkillCmd")
		//处理指定方向的技能的目标位置
		if skilldata.CastTargetType == 5 {
			dir := vec2d.Sub(vec2d.Vec2{float64(data.X), float64(data.Y)}, this.Body.Position)
			dir.Normalize()
			dir.MulToFloat64(float64(skilldata.CastRange + this.AddedMagicRange))
			dir.Collect(&this.Body.Position)
			this.SkillCmdData.X = float32(dir.X)
			this.SkillCmdData.Y = float32(dir.Y)
		}
	}

}

//检查技能指令的有效性
func (this *Unit) CheckSkillCmd() {

	if this.HaveSkillCmd() == false {
		return
	}

	///at := this.InScene.FindUnitByID(this.AttackCmdDataTarget.ID)
	if this.UseSkillEnable(this.SkillCmdData) == false {
		this.StopSkillCmd()
	}
}

//中断技能命令
func (this *Unit) StopSkillCmd() {
	this.SkillCmdData = nil
}

//-----------------------攻击命令--------------------
//是否有攻击命令
func (this *Unit) HaveAttackCmd() bool {

	if this.AttackCmdData != nil {
		return true
	}
	return false
}
func (this *Unit) CheckIsEnemy(target *Unit) bool {
	if target == nil {
		return false
	}
	//攻击模式(1:和平模式 2:组队模式 3:全体模式 4:阵营模式(玩家,NPC) 5:行会模式)
	if this.AttackMode == 1 {
		//阵营不同 //阵营(1:玩家 2:NPC)
		if this.Camp != target.Camp {
			return true
		} else {
			//如果目标是全体模式 也是敌人
			if target.AttackMode == 3 {
				return true
			}
		}
		return false
	} else if this.AttackMode == 3 {
		//全体模式下 不是同一个玩家控制的单位可以攻击
		if this.ControlID != target.ControlID {
			return true
		} else {
			if this.ControlID == -1 && this.ID != target.ID {
				return true
			}
			return false
		}

	}

	return false

}

//能否攻击(根据阵营,攻击模式判断与是否死亡)
func (this *Unit) CheckAttackEnable2Target(target *Unit) bool {

	if this.IsDisappear() || target.IsDisappear() || this.CanSeeTarget(target) == false || target.PhisicImmune == 1 {
		return false
	}

	//不能攻击
	if this.AttackEnable == 2 {
		return false
	}
	//检查是否是敌人
	if this.CheckIsEnemy(target) == false {
		return false
	}

	return true

}

//攻击行为命令
func (this *Unit) AttackCmd(data *protomsg.CS_PlayerAttack) {

	at := this.InScene.FindUnitByID(data.TargetUnitID)
	if at == nil {
		return
	}
	//判断阵营 攻击模式 是否允许本次攻击
	if this.CheckAttackEnable2Target(at) == true {
		this.AttackCmdData = data
		this.AttackCmdDataTarget = at
		//		if this.AttackMode == 3 {
		//			log.Info("---------AttackCmd")
		//		}

	}

}

//检查攻击指令的有效性 如果目标单位被场景删除 则无效
func (this *Unit) CheckAttackCmd() {

	if this.HaveAttackCmd() == false {
		return
	}
	if this.AttackCmdDataTarget == nil {
		return
	}
	///at := this.InScene.FindUnitByID(this.AttackCmdDataTarget.ID)
	if this.CheckAttackEnable2Target(this.AttackCmdDataTarget) == false {
		//log.Info("---------tttt:%d", this.AttackCmdDataTarget.ID)
		this.StopAttackCmd()
	}
}

//中断攻击命令
func (this *Unit) StopAttackCmd() {
	this.AttackCmdData = nil
	this.AttackCmdDataTarget = nil
	if this.UnitType == 1 {
		log.Info("---------StopAttackCmd")
	}

}

//-----------------------移动命令--------------------
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
	if this.MoveEnable == 2 {
		return false
	}

	return true
}

//设置单位朝向
func (this *Unit) SetDirection(dir vec2d.Vec2) {
	this.Body.Direction = dir
}

//移动行为命令
func (this *Unit) MoveCmd(data *protomsg.CS_PlayerMove) {
	if this.MoveEnable == 2 {
		//当不能移动的时候 检查是否可以转向
		if this.TurnEnable == 1 {
			v1 := vec2d.Vec2{X: float64(this.MoveCmdData.X), Y: float64(this.MoveCmdData.Y)}
			this.SetDirection(v1)
		}
		return
	}

	if this.AttackCmdData == nil {
		this.MoveCmdData = data
		return
	}
	//检测是否要中断攻击
	if this.MoveCmdData == nil {
		log.Info("---------111111")
		this.StopAttackCmd()
		this.MoveCmdData = data
	} else {
		if this.MoveCmdData.IsStart == false && data.IsStart == true {
			log.Info("---------2222")
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

	//log.Info("GetProjectileStartPos---:%f---:%f", this.ProjectileStartPosDis, this.ProjectileStartPosHeight)

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

//初始化hp和mp
func (this *Unit) InitHPandMP(hp float32, mp float32) {
	//满血 满蓝
	this.MAX_HP = this.BaseHP
	this.HP = int32(float32(this.MAX_HP) * hp)
	this.MAX_MP = this.BaseMP
	this.MP = int32(float32(this.MAX_MP) * mp)
	//log.Info("---hp:%d---mp:%d", this.HP, this.MP)
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

	NoCareDodge float32 //无视闪避几率

	MoveEnable   int32 //能否移动 (比如 被缠绕不能移动) 1:可以 2:不可以
	TurnEnable   int32 //能否转向 (比如 被眩晕不能转向) 1:可以 2:不可以
	AttackEnable int32 //能否攻击 (比如 被眩晕和缴械不能攻击) 1:可以 2:不可以
	SkillEnable  int32 //能否使用主动技能 (比如 被眩晕和沉默不能使用主动技能) 1:可以 2:不可以
	ItemEnable   int32 //能否使用主动道具 (比如 被眩晕和禁用道具不能使用主动道具) 1:可以 2:不可以
	MagicImmune  int32 //是否技能免疫 1：是 2:不是
	PhisicImmune int32 //是否物理攻击免疫 1:是 2:否

	AddedMagicRange float32 //额外施法距离
	ManaCost        float32 //魔法消耗降低 (0.1)表示降低 10%
	MagicCD         float32 //技能CD降低 (0.1)表示降低 10%

	Invisible int32 //隐身 1:是 2:否

	//强制移动相关
	ForceMoveRemainTime float32    //强制移动剩余时间
	ForceMoveSpeed      vec2d.Vec2 //强制移动速度 包括方向和大小
	ForceMoveLevel      int32      //强制移动等级

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

	InitPosition vec2d.Vec2 //初始位置

	Skills map[int32]*Skill  //所有技能
	Buffs  map[int32][]*Buff //所有buff 同typeID下可能有多个buff

	HaloInSkills map[int32][]int32 //来自被动技能的光环

	//每秒钟干事 剩余时间
	EveryTimeDoRemainTime float32 //每秒钟干事 的剩余时间

	//发送数据部分
	ClientData    *protomsg.UnitDatas //客户端显示数据
	ClientDataSub *protomsg.UnitDatas //客户端显示差异数据
}

func (this *Unit) SetAI(ai UnitAI) {
	this.AI = ai

}
func (this *Unit) FreshHaloInSkills() {
	for _, v := range this.Skills {
		if len(v.InitHalo) > 0 {
			//已经有此halo 就要删除之前的
			if _, ok := this.HaloInSkills[v.TypeID]; ok {
				for _, v1 := range this.HaloInSkills[v.TypeID] {
					this.InScene.RemoveHalo(v1)
				}
			}
			//---------------
			//log.Info("111111111111")
			halos := utils.GetInt32FromString2(v.InitHalo)
			re := make([]int32, 0)
			for _, v1 := range halos {
				halo := NewHalo(v1, v.Level)
				halo.SetParent(this)
				if halo != nil {
					this.InScene.AddHalo(halo)
					//log.Info("2222222222222")
				}
				re = append(re, halo.ID)
			}
			this.HaloInSkills[v.TypeID] = re

		}
	}
}

func CreateUnit(scene *Scene, typeid int32) *Unit {

	if scene == nil {
		return nil
	}

	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	//	文件数据
	unitre.UnitFileData = *(conf.GetUnitFileData(typeid))
	unitre.Name = unitre.UnitName
	unitre.Level = 1

	unitre.Init()
	unitre.InitHPandMP(1.0, 1.0)
	unitre.IsMain = 2
	//unitre.UnitType = 2 //单位类型(1:英雄 2:普通单位 3:远古 4:boss)
	unitre.ControlID = -1

	return unitre
}

func CreateUnitByPlayer(scene *Scene, player *Player, datas []byte) *Unit {
	if scene == nil || player == nil {
		return nil
	}
	unitre := &Unit{}
	unitre.ID = UnitID
	UnitID++
	unitre.InScene = scene
	unitre.MyPlayer = player

	//---------db.DB_CharacterInfo
	characterinfo := db.DB_CharacterInfo{}
	utils.Bytes2Struct(datas, &characterinfo)
	player.Characterid = characterinfo.Characterid

	log.Info("---DB_CharacterInfo---%v", characterinfo)
	//	文件数据
	unitre.UnitFileData = *(conf.GetUnitFileData(characterinfo.Typeid))
	unitre.InitHPandMP(characterinfo.HP, characterinfo.MP)

	//名字 等级 经验 创建时的位置
	unitre.Name = characterinfo.Name
	unitre.Level = characterinfo.Level
	unitre.Experience = characterinfo.Experience
	unitre.InitPosition = vec2d.Vec2{float64(characterinfo.X), float64(characterinfo.Y)}

	//创建技能
	skilldbdata := strings.Split(characterinfo.Skill, ";")
	unitre.Skills = NewUnitSkills(skilldbdata, unitre.InitSkillsInfo) //所有技能
	for _, v := range unitre.Skills {
		log.Info("-------new skill:%v", v)
	}
	//初始化技能被动光环
	unitre.HaloInSkills = make(map[int32][]int32)
	unitre.FreshHaloInSkills()

	//初始化
	unitre.Init()
	unitre.IsMain = 1
	unitre.ControlID = player.Uid

	return unitre
}

//初始化
func (this *Unit) Init() {
	this.State = NewIdleState(this)

	this.AttackMode = 1 //和平攻击模式
	this.EveryTimeDoRemainTime = 1

	this.IsDeath = 2

	this.ClearBuff()

	//弹道位置计算

	utils.GetFloat32FromString(this.ProjectileStartPos, &this.ProjectileStartPosDis, &this.ProjectileStartPosHeight)
	utils.GetFloat32FromString(this.ProjectileEndPos, &this.ProjectileEndPosDis, &this.ProjectileEndPosHeight)

	//this.TestData()
}

//设置强制移动相关
func (this *Unit) SetForceMove(time float32, speed vec2d.Vec2, level int32) {

	//direction.Normalize()

	if this.ForceMoveRemainTime <= 0 {
		this.ForceMoveRemainTime = time
		this.ForceMoveSpeed = speed
		this.ForceMoveLevel = level
	} else {
		if level >= this.ForceMoveLevel {
			this.ForceMoveRemainTime = time
			this.ForceMoveSpeed = speed
			this.ForceMoveLevel = level
		}
	}
}

//更新强制移动
func (this *Unit) UpdateForceMove(dt float64) {
	if this.ForceMoveRemainTime > 0 {
		this.ForceMoveRemainTime -= float32(dt)
		//movedelta := vec2d.Mul(this.ForceMoveSpeed.GetNormalized(),dt)
		this.Body.SpeedSize = this.ForceMoveSpeed.Length()
		this.Body.IsCollisoin = false
		this.Body.TurnDirection = false
		this.Body.CollisoinLevel = 2
		this.Body.SetMoveDir(this.ForceMoveSpeed)
	}
}

//
func (this *Unit) EveryTimeDo(dt float64) {

	if this.IsDisappear() {
		return
	}

	this.EveryTimeDoRemainTime -= float32(dt)
	if this.EveryTimeDoRemainTime <= 0 {
		//do
		this.EveryTimeDoRemainTime += 1

		//每秒回血
		this.ChangeHP(int32(this.HPRegain))
		this.ChangeMP(int32(this.MPRegain))
	}
}

func (this *Unit) ShowMiss(isshow bool) {
	if this.ClientData != nil {
		this.ClientData.IsMiss = isshow

	}
	if this.ClientDataSub != nil {
		this.ClientDataSub.IsMiss = isshow
		//		if isshow {
		//			log.Info("ShowMiss")
		//		}

	}
}

//
//更新 范围影响的buff 被动技能
func (this *Unit) PreUpdate(dt float64) {
	if this.IsDisappear() {
		return
	}
	this.ShowMiss(false)
}

//更新
func (this *Unit) Update(dt float64) {
	//设置是否有碰撞  设置移动速度 和逻辑状态

	//技能更新
	for _, v := range this.Skills {
		v.Update(dt)
	}
	//更新buff
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			if v1.IsEnd == true {
				//删除k1的元素
				v = append(v[:k1], v[k1+1:]...)
				if len(v) <= 0 {
					delete(this.Buffs, k)
				}
				continue
			}

			v1.Update(dt)
		}

	}

	//计算属性值
	this.CalProperty()

	//移动核心
	//移动行为逻辑速度设置
	this.Body.SpeedSize = this.MoveSpeed

	//AI
	if this.AI != nil {
		this.AI.Update(dt)
	}
	this.CheckSkillCmd()
	this.CheckAttackCmd()

	//
	this.EveryTimeDo(dt)

	//逻辑状态更新
	this.State.OnTransform()
	this.State.Update(dt)

	//强制移动更新
	this.UpdateForceMove(dt)

	//设置碰撞等级
	if this.IsDeath == 1 {
		this.Body.IsCollisoin = false
	}
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

}

//改变血量
func (this *Unit) ChangeHP(hp int32) int32 {
	lasthp := this.HP
	this.HP += hp
	if this.HP <= 0 {
		//死亡处理
		this.HP = 0
		this.IsDeath = 1
	} else {
		//死亡处理
		this.IsDeath = 2
	}
	if this.HP >= this.MAX_HP {
		this.HP = this.MAX_HP
	}
	return this.HP - lasthp
	//log.Info("---ChangeHP---:%d   :%d", hp, this.HP)
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

	//	if this.UnitType == 1 {
	//		log.Info("CalPhysicalAmaor %f   %f", this.PhysicalAmaor, this.PhysicalResist)
	//	}

}

//计算魔抗
func (this *Unit) CalMagicAmaor() {
	//非线性叠加
	//基础魔抗叠加力量带来的魔抗

	strenth := float32(this.AttributeStrength * conf.StrengthAddMagicAmaor)

	//装备
	//技能
	//buff

	this.MagicAmaor = utils.NoLinerAdd(this.BaseMagicAmaor, strenth)
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

	//计算无视闪避几率
	//装备
	//技能
	//buff
	this.NoCareDodge = 0
}

//计算生命回复
func (this *Unit) CalHPRegain() {
	this.HPRegain = this.BaseHPRegain + float32(this.AttributeStrength*conf.StrengthAddHPRegain)
	//装备
	//技能
	//buff
}

//计算异常状态
func (this *Unit) CalControlState() {
	this.MoveEnable = 1   //能否移动 (比如 被缠绕不能移动) 1:可以 2:不可以
	this.TurnEnable = 1   //能否转向 (比如 被眩晕不能转向) 1:可以 2:不可以
	this.AttackEnable = 1 //能否攻击 (比如 被眩晕和缴械不能攻击) 1:可以 2:不可以
	this.SkillEnable = 1  //能否使用主动技能 (比如 被眩晕和沉默不能使用主动技能) 1:可以 2:不可以
	this.ItemEnable = 1   //能否使用主动道具 (比如 被眩晕和禁用道具不能使用主动道具) 1:可以 2:不可以
	this.MagicImmune = 2
	this.PhisicImmune = 2

	this.AddedMagicRange = 0 //额外施法距离
	this.ManaCost = 0        //魔法消耗降低
	this.MagicCD = 0         //技能CD降低
	this.Invisible = 2       //隐身   否

	this.Body.IsCollisoin = true
	this.Body.TurnDirection = true
	this.Body.CollisoinLevel = 1
	this.Body.MoveDir = vec2d.Vec2{}
}

//计算单个buff对属性的影响
func (this *Unit) CalPropertyByBuffCR(v1 *Buff) {
	if v1 == nil || v1.IsActive == false {
		return
	}
	this.Attack += int32(float32(this.Attack) * v1.AttackCR)
	if this.Attack < 0 {
		this.Attack = 0
	}
	this.MoveSpeed += this.MoveSpeed * float64(v1.MoveSpeedCR)
	if this.MoveSpeed < 0 {
		this.MoveSpeed = 0
	}
	this.MPRegain += this.MPRegain * v1.MPRegainCR
	if this.MPRegain < 0 {
		this.MPRegain = 0
	}
	this.PhysicalAmaor += this.PhysicalAmaor * v1.PhysicalAmaorCR
	if this.PhysicalAmaor < 0 {
		this.PhysicalAmaor = 0
	}
	this.HPRegain += this.HPRegain * v1.HPRegainCR
	if this.HPRegain < 0 {
		this.HPRegain = 0
	}

}

//计算单个buff对属性的影响
func (this *Unit) CalPropertyByBuffCV(v1 *Buff) {
	if v1 == nil || v1.IsActive == false {
		return
	}

	//log.Info("--11--speed:%f", this.AttackSpeed)
	this.AttributeStrength += v1.AttributeStrengthCV
	this.AttributeIntelligence += v1.AttributeIntelligenceCV
	this.AttributeAgility += v1.AttributeAgilityCV
	this.AttackSpeed += v1.AttackSpeedCV

	this.Attack += int32(v1.AttackCV)

	this.MoveSpeed += float64(v1.MoveSpeedCV)
	this.MagicScale = utils.NoLinerAdd(this.MagicScale, v1.MagicScaleCV)

	this.MPRegain += v1.MPRegainCV

	this.PhysicalAmaor += v1.PhysicalAmaorCV
	this.MagicAmaor = utils.NoLinerAdd(this.MagicAmaor, v1.MagicAmaorCV)
	this.Dodge = utils.NoLinerAdd(this.Dodge, v1.DodgeCV)

	this.HPRegain += v1.HPRegainCV
	this.HPRegain += v1.HPRegainCVOfMaxHP * float32(this.MAX_HP)

	this.NoCareDodge = utils.NoLinerAdd(this.NoCareDodge, v1.NoCareDodgeCV)
	this.AddedMagicRange += v1.AddedMagicRangeCV
	this.ManaCost = utils.NoLinerAdd(this.ManaCost, v1.ManaCostCV)
	this.MagicCD = utils.NoLinerAdd(this.MagicCD, v1.MagicCDCV)

	if v1.NoMove == 1 {
		this.MoveEnable = 2
	}
	if v1.NoTurn == 1 {
		this.TurnEnable = 2
	}
	if v1.NoAttack == 1 {
		this.AttackEnable = 2
	}
	if v1.NoSkill == 1 {
		this.SkillEnable = 2
	}
	if v1.NoItem == 1 {
		this.ItemEnable = 2
	}
	if v1.MagicImmune == 1 {
		this.MagicImmune = 1
	}
	if v1.Invisible == 1 {
		this.Invisible = 1
	}
	if v1.PhisicImmune == 1 {
		this.PhisicImmune = 1
	}

	//log.Info("--22--speed:%f", this.AttackSpeed)
}

//计算所有buff对属性的影响
func (this *Unit) CalPropertyByBuffs() {
	//技能携带的buf cr
	for _, v := range this.Skills {
		if v.Level <= 0 {
			continue
		}

		buffs := utils.GetInt32FromString2(v.InitBuff)
		for _, v1 := range buffs {
			buff := NewBuff(v1, v.Level, this)
			if buff != nil {
				this.CalPropertyByBuffCR(buff)
			}
		}
	}
	//buff cr
	for _, v := range this.Buffs {
		for _, v1 := range v {
			this.CalPropertyByBuffCR(v1)

		}

	}

	//技能携带的buf cv
	for _, v := range this.Skills {
		if v.Level <= 0 {
			continue
		}

		buffs := utils.GetInt32FromString2(v.InitBuff)
		for _, v1 := range buffs {
			buff := NewBuff(v1, v.Level, this)
			if buff != nil {
				this.CalPropertyByBuffCV(buff)
			}
		}
	}
	//buff cv
	for _, v := range this.Buffs {
		for _, v1 := range v {
			this.CalPropertyByBuffCV(v1)
		}

	}

	//攻击速度取值范围
	if this.AttackSpeed <= 10 {
		this.AttackSpeed = 10
	} else if this.AttackSpeed >= float32(this.BaseMaxAttackSpeed) {
		this.AttackSpeed = float32(this.BaseMaxAttackSpeed)
	}
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
	//计算异常状态
	this.CalControlState()

	//计算buff对属性的影响
	this.CalPropertyByBuffs()
}

//清空buff
func (this *Unit) ClearBuff() {
	this.Buffs = make(map[int32][]*Buff)
}

//删除buff 删除使用技能后失效的buff
func (this *Unit) RemoveBuffForDoSkilled() {
	//buff
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			if v1.DoSkilledInvalid == 1 {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}
		}
	}
}

//删除buff 删除攻击后失效的buff
func (this *Unit) RemoveBuffForAttacked() {
	//buff
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			if v1.AttackedInvalid == 1 {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}

		}

	}
}

//被目标 驱散自己的buff
func (this *Unit) ClearBuffForTarget(target *Unit, clearlevel int32) {
	if target == nil || clearlevel <= 0 || target.IsDisappear() {
		return
	}

	//buff
	isenemy := target.CheckIsEnemy(this)
	for k, v := range this.Buffs {
		for k1, v1 := range v {
			//BuffType         int32 //buff类型 1:表示良性 2:表示恶性  队友只能驱散我的恶性buff 敌人只能驱散我的良性buff
			//ClearLevel       int32 //驱散等级 1 表示需要驱散等级大于等于1的 驱散效果才能驱散此buff
			if isenemy == true && v1.BuffType == 1 && clearlevel >= v1.ClearLevel {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}
			if isenemy == false && v1.BuffType == 2 && clearlevel >= v1.ClearLevel {
				this.Buffs[k] = append(this.Buffs[k][:k1], this.Buffs[k][k1+1:]...)
			}

		}

	}

}

//通过 buff 添加buff
func (this *Unit) AddBuffFromBuff(buff *Buff, castunit *Unit) *Buff {

	if castunit == nil || castunit.IsDisappear() || this.IsDisappear() {
		return nil
	}
	//攻击距离类型
	if buff.ActiveUnitAcpabilities == 1 && this.AttackAcpabilities != 1 {
		return nil
	}
	if buff.ActiveUnitAcpabilities == 2 && this.AttackAcpabilities != 2 {
		return nil
	}
	//BuffType         int32 //buff类型 1:表示良性 2:表示恶性  队友只能驱散我的恶性buff 敌人只能驱散我的良性buff
	isenemy := castunit.CheckIsEnemy(this)
	//如果是敌人 且 是良性buff 就不添加
	if isenemy == true && buff.BuffType == 1 {
		return nil
	}
	//如果不是敌人 且 是恶性buff 就不添加
	if isenemy == false && buff.BuffType == 2 {
		return nil
	}

	//如果恶性buff 单位魔法免疫 buff没有无视技能免疫
	if buff.BuffType == 2 && this.MagicImmune == 1 && buff.NoCareMagicImmuneAddBuff == 2 {
		return nil
	}

	bf, ok := this.Buffs[buff.TypeID]
	//叠加机制
	//		OverlyingType          int32 //叠加类型 1:只更新最大时间 2:完美叠加(小鱼的偷属性)
	//	OverlyingAddTag        int32 //叠加时是否增加标记数字 1:表示增加 2:表示不增加
	if ok == true && len(bf) > 0 {
		if buff.OverlyingType == 1 {

			if bf[0].RemainTime < buff.Time {
				bf[0].RemainTime = buff.Time
			}
			return bf[0]

		} else if buff.OverlyingType == 2 {
			bf = append(bf, buff)
			return buff
		}
	} else {
		bfs := make([]*Buff, 0)
		bfs = append(bfs, buff)
		this.Buffs[buff.TypeID] = bfs
		//给单位计算buff效果
		this.CalPropertyByBuffCV(buff)

		return buff
	}
	return nil
}

//通过bufftypeid string 添加buff  castunit给我添加
func (this *Unit) AddBuffFromStr(buffsstr string, level int32, castunit *Unit) []*Buff {
	buffs := utils.GetInt32FromString2(buffsstr)
	re := make([]*Buff, 0)
	for _, v := range buffs {
		buff := NewBuff(v, level, this)
		if buff != nil {
			buff = this.AddBuffFromBuff(buff, castunit)
			re = append(re, buff)
		}
	}
	return re
}

//通过bufftypeid string 添加halo
func (this *Unit) AddHaloFromStr(halosstr string, level int32, pos *vec2d.Vec2) []*Halo {
	halos := utils.GetInt32FromString2(halosstr)
	re := make([]*Halo, 0)
	for _, v := range halos {
		halo := NewHalo(v, level)
		halo.SetParent(this)
		if pos != nil {
			halo.Position = *pos
		}
		if halo != nil {
			this.InScene.AddHalo(halo)
		}
		re = append(re, halo)
	}
	return re
}

//受到来自子弹的伤害
func (this *Unit) BeAttacked(bullet *Bullet) int32 {
	//计算闪避
	if bullet.SkillID == -1 {
		//普通攻击
		isDodge := false //闪避
		//无视闪避
		if utils.CheckRandom(this.NoCareDodge) {
			isDodge = false
		} else {
			if utils.CheckRandom(this.Dodge) {
				isDodge = true
			} else {
				isDodge = false
			}
		}

		//闪避了
		if isDodge {
			//本单位显示miss
			this.ShowMiss(true)
			return 0
		}
	}
	//计算伤害
	physicAttack := int32(0)
	if this.PhisicImmune != 1 {
		physicAttack = bullet.GetAttackOfType(1) //物理攻击
		//计算护甲抵消后伤害
		physicAttack = int32(utils.SetValueGreaterE(float32(physicAttack)*(1-this.PhysicalResist), 0))
	}
	magicAttack := bullet.GetAttackOfType(2) //魔法攻击
	pureAttack := bullet.GetAttackOfType(3)  //纯粹攻击

	//计算魔抗抵消后伤害
	magicAttack = int32(utils.SetValueGreaterE(float32(magicAttack)*(1-this.MagicAmaor), 0))
	//magicAttack = utils.SetValueGreaterE(magicAttack,0)

	//-----扣血--
	hurtvalue := -(physicAttack + magicAttack + pureAttack)
	this.ChangeHP(hurtvalue)
	//log.Info("---hurtvalue---%d   %f", hurtvalue, this.PhysicalResist)
	return hurtvalue
}

//创建子弹
func (this *Unit) AddBullet(bullet *Bullet) {
	if this.InScene != nil {
		this.InScene.AddBullet(bullet)
	}
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
	this.ClientData.IsDeath = this.IsDeath
	this.ClientData.Invisible = this.Invisible
	this.ClientData.Camp = this.Camp

	//技能
	this.ClientData.SD = make([]*protomsg.SkillDatas, 0)
	for _, v := range this.Skills {
		skdata := &protomsg.SkillDatas{}
		skdata.TypeID = v.TypeID
		skdata.Level = v.Level
		skdata.RemainCDTime = v.RemainCDTime
		skdata.CanUpgrade = int32(2) //v.CanUpgrade
		skdata.Index = v.Index
		skdata.CastType = v.CastType
		skdata.CastTargetType = v.CastTargetType
		skdata.UnitTargetTeam = v.UnitTargetTeam
		skdata.UnitTargetCamp = v.UnitTargetCamp
		skdata.NoCareMagicImmune = v.NoCareMagicImmune
		skdata.CastRange = v.CastRange + this.AddedMagicRange
		skdata.Cooldown = v.Cooldown
		skdata.HurtRange = v.HurtRange
		skdata.ManaCost = v.ManaCost
		skdata.AttackAutoActive = v.AttackAutoActive
		this.ClientData.SD = append(this.ClientData.SD, skdata)
	}
	//Buffs  map[int32][]*Buff //所有buff 同typeID下可能有多个buff
	this.ClientData.BD = make([]*protomsg.BuffDatas, 0)
	for _, v := range this.Buffs {
		if len(v) <= 0 {
			continue
		}
		buffdata := &protomsg.BuffDatas{}
		buffdata.TypeID = v[0].TypeID
		buffdata.RemainTime = v[0].RemainTime
		buffdata.Time = v[0].Time
		if len(v) > 1 {
			buffdata.TagNum = int32(len(v))
		} else {
			buffdata.TagNum = v[0].TagNum
		}
		this.ClientData.BD = append(this.ClientData.BD, buffdata)
	}

	//Skills map[int32]*Skill //所有技能

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

	//当前数据与上一次数据对比 相减 数值部分
	this.ClientDataSub.HP = this.HP - this.ClientData.HP
	this.ClientDataSub.MaxHP = this.MAX_HP - this.ClientData.MaxHP
	this.ClientDataSub.MP = this.MP - this.ClientData.MP
	this.ClientDataSub.MaxMP = this.MAX_MP - this.ClientData.MaxMP
	this.ClientDataSub.Level = this.Level - this.ClientData.Level
	this.ClientDataSub.Experience = this.Experience - this.ClientData.Experience
	this.ClientDataSub.MaxExperience = this.MaxExperience - this.ClientData.MaxExperience
	this.ClientDataSub.ControlID = this.ControlID - this.ClientData.ControlID

	this.ClientDataSub.AnimotorState = this.AnimotorState - this.ClientData.AnimotorState
	this.ClientDataSub.AttackTime = this.GetOneAttackTime() - this.ClientData.AttackTime

	this.ClientDataSub.X = float32(this.Body.Position.X) - this.ClientData.X
	this.ClientDataSub.Y = float32(this.Body.Position.Y) - this.ClientData.Y

	this.ClientDataSub.DirectionX = float32(this.Body.Direction.X) - this.ClientData.DirectionX
	this.ClientDataSub.DirectionY = float32(this.Body.Direction.Y) - this.ClientData.DirectionY

	this.ClientDataSub.UnitType = this.UnitType - this.ClientData.UnitType
	this.ClientDataSub.AttackAcpabilities = this.AttackAcpabilities - this.ClientData.AttackAcpabilities
	this.ClientDataSub.AttackMode = this.AttackMode - this.ClientData.AttackMode

	this.ClientDataSub.IsMain = this.IsMain - this.ClientData.IsMain
	this.ClientDataSub.IsDeath = this.IsDeath - this.ClientData.IsDeath
	this.ClientDataSub.Invisible = this.Invisible - this.ClientData.Invisible
	this.ClientDataSub.Camp = this.Camp - this.ClientData.Camp

	//技能
	this.ClientDataSub.SD = make([]*protomsg.SkillDatas, 0)
	for _, v := range this.Skills {
		skdata := &protomsg.SkillDatas{}
		skdata.TypeID = v.TypeID
		//上次发送的数据
		lastdata := &protomsg.SkillDatas{}
		for _, v1 := range this.ClientData.SD {
			if v1.TypeID == v.TypeID {
				lastdata = v1
				break
			}
		}

		skdata.Level = v.Level - lastdata.Level
		skdata.RemainCDTime = v.RemainCDTime - lastdata.RemainCDTime
		skdata.CanUpgrade = int32(2) - lastdata.CanUpgrade //v.CanUpgrade
		skdata.Index = v.Index - lastdata.Index
		skdata.CastType = v.CastType - lastdata.CastType
		skdata.CastTargetType = v.CastTargetType - lastdata.CastTargetType
		skdata.UnitTargetTeam = v.UnitTargetTeam - lastdata.UnitTargetTeam
		skdata.UnitTargetCamp = v.UnitTargetCamp - lastdata.UnitTargetCamp
		skdata.NoCareMagicImmune = v.NoCareMagicImmune - lastdata.NoCareMagicImmune
		skdata.CastRange = v.CastRange + this.AddedMagicRange - lastdata.CastRange
		skdata.Cooldown = v.Cooldown - lastdata.Cooldown
		skdata.HurtRange = v.HurtRange - lastdata.HurtRange
		skdata.ManaCost = v.ManaCost - lastdata.ManaCost
		skdata.AttackAutoActive = v.AttackAutoActive - lastdata.AttackAutoActive
		this.ClientDataSub.SD = append(this.ClientDataSub.SD, skdata)
	}

	this.ClientDataSub.BD = make([]*protomsg.BuffDatas, 0)
	for _, v := range this.Buffs {
		if len(v) <= 0 {
			continue
		}
		buffdata := &protomsg.BuffDatas{}
		buffdata.TypeID = v[0].TypeID
		//上次发送的数据
		lastdata := &protomsg.BuffDatas{}
		for _, v1 := range this.ClientData.BD {
			if v1.TypeID == v[0].TypeID {
				lastdata = v1
				break
			}
		}

		buffdata.RemainTime = v[0].RemainTime - lastdata.RemainTime
		buffdata.Time = v[0].Time - lastdata.Time
		if len(v) > 1 {
			buffdata.TagNum = int32(len(v)) - lastdata.TagNum
		} else {
			buffdata.TagNum = v[0].TagNum - lastdata.TagNum
		}

		this.ClientDataSub.BD = append(this.ClientDataSub.BD, buffdata)
	}

}

//被删除的时候
func (this *Unit) OnDestroy() {
	this.IsDelete = true
}

//即时属性获取
