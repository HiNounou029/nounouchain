// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package bridge

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type PlainChainInfo struct {
	Name         string
	NetworkID    int
	Endpoint     string
	ContractAddr string
}

type ConfigInfo struct {
	ABIFilePath   string
	ServerKeyPath string
	ClientKeyPath string
	Chains        []PlainChainInfo
}

func ReadConfig(configFilePath string) (*ConfigInfo, error) {
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, fmt.Errorf("read file %s: %v", configFilePath, err)
	}
	config := ConfigInfo{}
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("unmarshal data %s: %v", data, err)
	}

	return &config, nil
}

func WriteConfig() {
	config := ConfigInfo{
		ABIFilePath:   "/Volumes/sys/cefa/src/github.com/HiNounou029/nounouchain/cmd/bridge/contract/bridge.json",
		ServerKeyPath: "/Users/majunchang/sys/cefa/bin/account1.key",
		ClientKeyPath: "/Users/majunchang/sys/cefa/bin/account1.key",
		Chains: []PlainChainInfo{
			{Name: "chain1", NetworkID: 1,
				Endpoint:     "http://127.0.0.1:8669",
				ContractAddr: "0x36299d7c4e3c2d329b7cc6ec6e3ad9eab8d980cd"},

			{Name: "chain2", NetworkID: 100,
				Endpoint:     "http://cefa_1:8669",
				ContractAddr: "0xc5ea599b271e679ab2b8163663c30ad28dd4229f"},
		}}

	strConfig, _ := json.MarshalIndent(config, "", "\t")
	fmt.Printf("%s", strConfig)
}
