// Copyright 2014 mqant Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conf

import (
	"dq/log"
	"dq/utils"
)

var (
	UnitFileDatas = make(map[interface{}]interface{})

	//每一点主属性增加攻击力
	AttributePrimaryAddAttack float32 = 1

	//每一点力量增加血量上限 白字
	StrengthAddHP float32 = 20
	//每一点力量增加 每秒回血量 白字
	StrengthAddHPRegain float32 = 0.1
	//每一点力量增加 魔法抗性 白字
	StrengthAddMagicAmaor float32 = 0.0008

	//每一点智力增加MP上限 白字
	IntelligenceAddMP float32 = 12
	//每一点智力增加 每秒回蓝量 白字
	IntelligenceAddMPRegain float32 = 0.05
	//每一点智力增加 技能增强 白字
	IntelligenceAddMagicScale float32 = 0.0007

	//每一点敏捷增加 护甲 白字
	AgilityAddPhysicalAmaor float32 = 0.16111
	//每一点敏捷增加 攻击速度 白字
	AgilityAddAttackSpeed float32 = 1
	//每一点敏捷增加 移动速度比率 绿字
	AgilityAddMoveSpeed float32 = 0.0005
)

//场景配置文件
func LoadUnitFileData() {
	_, UnitFileDatas = utils.ReadXlsxData("bin/conf/units.xlsx", (*UnitFileData)(nil))
	//	for k, v := range UnitFileDatas {
	//		log.Info("data:%d %v", k, v)
	//	}
}
func GetUnitFileData(typeid int32) *UnitFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (UnitFileDatas[int(typeid)])
	if re == nil {
		log.Info("not find unitfile:%d", typeid)
		return nil
	}
	return (UnitFileDatas[int(typeid)]).(*UnitFileData)
}

//单位配置文件数据
type UnitFileData struct {
	//配置文件数据
	TypeID                    int32   //类型ID
	UnitName                  string  //单位名字
	ModeType                  string  //模型
	BaseHP                    int32   //基础HP
	BaseMP                    int32   //基础MP
	BaseAttackSpeed           int32   //基础攻击速度(141点攻击速度等于 1.20秒一次)
	BaseMaxAttackSpeed        int32   //基础最大攻击速度
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
	ProjectileStartPos        string  //弹道起始点 1,0.5  1表示单位正前方1米0.5表示单位高度0.5米的位置
	ProjectileEndPos          string  //弹道结束点 0,0.5  0表示单位正前方1米0.5表示单位高度0.5米的位置
	UnitType                  int32   //单位类型(1:英雄 2:普通单位 3:远古 4:boss)
	AttackAcpabilities        int32   //(1:近程攻击 2:远程攻击)

	//-------------新加----
	AutoAttackTraceRange    float32 //自动攻击的追击范围
	AutoAttackTraceOutRange float32 //自动攻击的取消追击范围
	//-----
	Camp int32 //阵营(1:玩家 2:NPC)  玩家的召唤物和幻象camp也是 玩家

	InitSkillsInfo   string  //初始技能信息 逗号分隔
	CollisionR       float64 //碰撞半径
	Death2RemoveTime float64 //死亡到删除的时间
}
