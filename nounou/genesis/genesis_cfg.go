package genesis

import (
	"encoding/json"
	"github.com/inconshreveable/log15"
	"io/ioutil"
)

var log = log15.New("pkg", "genesis")

type Account struct {
	Address string
	Id      string
}

type Config struct {
	Authorities []*Account //打包block
	Approvers   []*Account //预分配tokens, approve authority
}

// default configuration, should read from config file, e.g., /data/genesis_cfg.json
var genesisCfg = `
{
  "Authorities": [
    {
      "Address": "0xe79c2f9aa58f8fef756c9c26ffa3fe6c6d7964dd",
      "Id": "0x0000000000000000000000000000000000000000000000000000006e6f646531"
    },
    {
      "Address": "0xc7c6775bdb39af5ee200400def9a94f4460479a5",
      "Id": "0x0000000000000000000000000000000000000000000000000000006e6f646532"
    },
    {
      "Address": "0x130980d09a5185c0e6286713a180a44abba6379d",
      "Id": "0x0000000000000000000000000000000000000000000000000000006e6f646533"
    },
    {
      "Address": "0x79533f2a910b65a3a8279265d2966c90dfc1075c",
      "Id": "0x0000000000000000000000000000000000000000000000000000006e6f646534"
    },
    {
      "Address": "0xeb909c29b7b526e94060f3dacc03e2fd89fa05e3",
      "Id": "0x0000000000000000000000000000000000000000000000000000006e6f646535"
    },
    {
      "Address": "0xf51a27e8871ba426489c42df6835c143167472a1",
      "Id": "0x0000000000000000000000000000000000000000000000000000006e6f646536"
    },
    {
      "Address": "0xedcc8e209411277a547d774bd3eddc65efbf09fa",
      "Id": "0x0000000000000000000000000000000000000000000000000000006e6f646537"
    }
  ],
  "Approvers": [
    {
      "Address": "0x875b8d9fe06b5deab8ed89cce0e66b4616effbac",
      "Id": "0x00000000000000000000000000000000000000000000617070726f7665725f31"
    },
    {
      "Address": "0x6a9e8f20ff77334cb4d6b5d3cc57bb855c0e6e6c",
      "Id": "0x00000000000000000000000000000000000000000000617070726f7665725f32"
    },
    {
      "Address": "0x99e347d24ce8cf38e22299341f5e9bd302312eb0",
      "Id": "0x00000000000000000000000000000000000000000000617070726f7665725f33"
    }
  ]
}

`

func MustReadConfig(filePath string) *Config {
	cfgBytes := []byte(genesisCfg)
	if len(filePath) > 0 {
		bytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Info("read genesis config file error, using default configuration",
				"err", err)
		} else {
			cfgBytes = bytes
		}
	}
	cfg := &Config{}
	if err := json.Unmarshal(cfgBytes, cfg); err != nil {
		panic(err)
	}
	return cfg
}
