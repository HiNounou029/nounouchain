// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package bridge

import (
	"encoding/json"
	"testing"
)

func TestReadConfig(t *testing.T) {
	configFilePath := "/Volumes/sys/cefa/src/github.com/HiNounou029/nounouchain/cmd/bridge/config.json"
	cfg, err := ReadConfig(configFilePath)
	if err != nil {
		t.Logf("%v", err)
		return
	}
	jsCfg, _ := json.MarshalIndent(cfg, "", "\t")

	t.Logf("%s", jsCfg)
}
