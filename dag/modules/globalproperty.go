/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/
/*
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 *
 */

package modules

import (
	"time"

	"github.com/palletone/go-palletone/core"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common"
)

// 全局属性的结构体定义
type GlobalProperty struct {
	ChainParameters core.ChainParameters // 区块链网络参数

	ActiveMediators map[Mediator]bool // 当前活跃mediator集合；每个维护间隔更新一次
}

// 动态全局属性的结构体定义
type DynamicGlobalProperty struct {
	LastVerifiedUnitNum uint32 // 最近的验证单元编号(数量)

	//	VerifiedUnitHash string // 最近的验证单元hash

	//	LastVerifiedUnit *v.VerifiedUnit	// 最近生产的验证单元

	LastVerifiedUnitTime time.Time // 最近的验证单元时间

	//	CurrentMediator *Mediator // 当前生产验证单元的mediator, 用于判断是否连续同一个mediator生产验证单元

	//	NextMaintenanceTime time.Time // 下一次系统维护时间

	// 当前的绝对时间槽数量，== 从创世开始所有的时间槽数量 == verifiedUnitNum + 丢失的槽数量
	CurrentASlot uint64

	/**
	在过去的128个见证单元生产slots中miss的数量。
	The count of verifiedUnit production slots that were missed in the past 128 verifiedUnits
	用于计算mediator的参与率。used to compute mediator participation.
	*/
	//	RecentSlotsFilled float32
}

func NewGlobalProp() (*GlobalProperty) {
	return &GlobalProperty{
		ChainParameters: core.NewChainParams(),
		ActiveMediators: map[Mediator]bool{},
	}
}

func NewDynGlobalProp() (*DynamicGlobalProperty) {
	return &DynamicGlobalProperty{
		LastVerifiedUnitNum: 0,
		CurrentASlot:        0,
	}
}

func InitGlobalProp(genesis *core.Genesis) (*GlobalProperty) {
	log.Info("initialize global property...")

	// Create global properties
	gp := NewGlobalProp()

	log.Info("initialize chain parameters...")
	gp.ChainParameters = genesis.InitialParameters

	log.Info("Set active mediators...")
	// Set active mediators
	for i := uint16(0); i < genesis.InitialActiveMediators; i++ {
		ad := common.StringToAddress(genesis.InitialMediatorCandidates[i])
		md := Mediator{
			Address:ad,
		}
		gp.ActiveMediators[md] = true
	}

	return gp
}

func InitDynGlobalProp(genesis *core.Genesis) (*DynamicGlobalProperty) {
	log.Info("initialize dynamic global property...")

	// Create dynamic global properties
	dgp := NewDynGlobalProp()
	dgp.LastVerifiedUnitTime = time.Unix(genesis.InitialTimestamp, 0)

	return dgp
}
