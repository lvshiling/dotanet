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
	"dq/protobuf"
	"dq/utils"
)

var (
	StoreFileDatas = make(map[interface{}]interface{})
)

//场景配置文件
func LoadStoreFileData() {
	_, StoreFileDatas = utils.ReadXlsxData("bin/conf/store.xlsx", (*CommodityData)(nil))

}

func GetStoreFileData(typeid int32) *CommodityData {
	//log.Info("find unitfile:%d", typeid)

	re := (StoreFileDatas[int(typeid)])
	if re == nil {
		log.Info("not find StoreFileDatas:%d", typeid)
		return nil
	}
	return (StoreFileDatas[int(typeid)]).(*CommodityData)
}

func GetStoreData2SC_StoreData() *protomsg.SC_StoreData {
	re := &protomsg.SC_StoreData{}
	re.Commoditys = make([]*protomsg.CommodityDataProto, 0)
	for _, v := range StoreFileDatas {
		one := &protomsg.CommodityDataProto{}
		if v.(*CommodityData).IsSell != 1 {
			continue
		}
		one.TypeID = v.(*CommodityData).TypeID
		one.ItemID = v.(*CommodityData).ItemID
		one.PriceType = v.(*CommodityData).PriceType
		one.Price = v.(*CommodityData).Price
		one.Level = v.(*CommodityData).Level
		re.Commoditys = append(re.Commoditys, one)
	}

	return re
}

//单位配置文件数据
type CommodityData struct {
	//配置文件数据
	TypeID    int32 //商品ID
	ItemID    int32 //道具ID
	Level     int32 //道具等级
	PriceType int32 //价格类型 1金币 2砖石
	Price     int32 //价格
	IsSell    int32 //是否售卖 1:卖  2:不卖
}
