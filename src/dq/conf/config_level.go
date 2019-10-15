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
	LevelFileDatas = make(map[interface{}]interface{})

	MaxLevel = int32(20)
)

//场景配置文件
func LoadLevelFileData() {
	_, LevelFileDatas = utils.ReadXlsxData("bin/conf/level.xlsx", (*LevelFileData)(nil))

}
func GetLevelFileData(level int32) *LevelFileData {
	//log.Info("find unitfile:%d", typeid)

	re := (LevelFileDatas[int(level)])
	if re == nil {
		log.Info("not find LevelFileDatas:%d", level)
		return nil
	}
	return (LevelFileDatas[int(level)]).(*LevelFileData)
}

//单位配置文件数据
type LevelFileData struct {
	//配置文件数据
	Level               int32   //等级
	UpgradeExperience   int32   //升级所需要的经验
	MaxExperienceOneDay int32   //一天中能获取到的最大经验值
	ReviveTime          float32 //复活时间
}
