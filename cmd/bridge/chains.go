// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package bridge

import (
	"fmt"
	"github.com/HiNounou029/nounouchain/poloclient"
	"github.com/HiNounou029/nounouchain/cmd/bridge/contract"
	"io/ioutil"
	"math/big"
	"strconv"
)

// chain information
type ChainInfo struct {
	Name         string
	NetworkID    *big.Int
	Endpoint     string
	ContractAddr poloclient.Address
	Bridge       *contract.BridgeContract
}

type ChainMap struct {
	chainMap map[string]*ChainInfo
}

func (cm *ChainMap) Chains() []*ChainInfo {
	chains := make([]*ChainInfo, 0, len(cm.chainMap))
	for _, v := range cm.chainMap {
		chains = append(chains, v)
	}
	return chains
}

func (cm *ChainMap) ChainIds() []*big.Int {
	chainIds := make([]*big.Int, 0, len(cm.chainMap))
	for _, v := range cm.chainMap {
		chainIds = append(chainIds, v.NetworkID)
	}
	return chainIds
}

func (cm *ChainMap) ById(id *big.Int) *ChainInfo {
	return cm.chainMap[id.String()]
}

func (cm *ChainMap) ByInt(id int) *ChainInfo {
	return cm.chainMap[strconv.Itoa(id)]
}

func newChainInfo(cfg *ConfigInfo, plain *PlainChainInfo, client bool) *ChainInfo {
	chain := &ChainInfo{
		Name:         plain.Name,
		NetworkID:    big.NewInt(int64(plain.NetworkID)),
		Endpoint:     plain.Endpoint,
		ContractAddr: poloclient.MustParseAddress(plain.ContractAddr),
	}
	keyPath := cfg.ServerKeyPath
	if client {
		keyPath = cfg.ClientKeyPath
	}
	bc, _ := poloclient.NewClient(chain.Endpoint, keyPath)
	abiBytes, err := ioutil.ReadFile(cfg.ABIFilePath)
	if err != nil {
		panic(fmt.Sprintf("open file %s: %v", cfg.ABIFilePath, err))
	}
	contract, err := contract.NewBridgeContract(bc, abiBytes, chain.ContractAddr)
	if err != nil {
		panic(fmt.Sprintf("new contract err %v", err))
	}
	chain.Bridge = contract

	return chain
}

// client: 作为普通client，还是bridge server, bridge server设置为false
func InitChains(configFilePath string, client bool) (*ChainMap, error) {
	cfg, err := ReadConfig(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %v", configFilePath, err)
	}
	var cm = &ChainMap{
		chainMap: make(map[string]*ChainInfo),
	}

	for _, chain := range cfg.Chains {
		cm.chainMap[strconv.Itoa(chain.NetworkID)] = newChainInfo(cfg, &chain, client)
	}
	return cm, nil
}
