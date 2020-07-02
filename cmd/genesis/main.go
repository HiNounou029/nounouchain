package main

import (
	"os"
	"fmt"
	"gopkg.in/urfave/cli.v1"
	"path/filepath"
	"github.com/HiNounou029/polo-sdk-go/accounts/keystore"
	"github.com/HiNounou029/polo-sdk-go/cmd/utils"
	"encoding/json"
	"io/ioutil"
	utils2 "github.com/HiNounou029/nounouchain/cmd/utils"
)

func main() {
	app := cli.App{
		Version:   "1.0",
		Name:      "genesis",
		Usage:     "generate genesis configuration files",
		Copyright: "2019 PoloChain",
		Flags:     []cli.Flag{
			DataDirFlag,
			ApproverCountFlag,
			AuthCountFlag,
			PasswordFileFlag,
		},
		Action:    genesis,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type entry struct {
	Address string
	Id string
}

type config struct {
	Authorities []*entry //打包block
	Approvers   []*entry //预分配tokens, approve authority
}

func genesis(ctx *cli.Context) error {

	scryptN := keystore.StandardScryptN
	scryptP := keystore.StandardScryptP

	dataDir := ctx.GlobalString(DataDirFlag.Name)
	if len(dataDir) == 0 {
		utils.Fatalf("genesis data dir must be given as argument")
	}

	if filepath.IsAbs(dataDir) == false {
		dir, err := filepath.Abs(dataDir)
		if err != nil {
			utils.Fatalf(err.Error())
		}

		dataDir = dir
	}

	authCount := ctx.GlobalInt(AuthCountFlag.Name)
	if authCount < 1 || authCount > 25 {
		utils.Fatalf("Authority count of genesis config, range[1,25]")
	}

	approverCount := ctx.GlobalInt(ApproverCountFlag.Name)
	if approverCount < 1 || approverCount > 25 {
		utils.Fatalf("Approver count of genesis config, range[1,25]")
	}

	utils.Logf("Authority count: %d", authCount)
	utils.Logf("Approver count: %d", approverCount)

	genConfig := &config{}
	genConfig.Authorities = make([]*entry, authCount)
	genConfig.Approvers = make([]*entry, approverCount)

	authIdPrefix :=     "0x0000000000000000000000000000000000000000000000000000006e6f6465"
	approverIdPrefix := "0x00000000000000000000000000000000000000000000617070726f7665725f"


	utils.Logf("Genesis authorities keystore file")
	for i := 0; i < authCount; i++ {
		newDir := filepath.Join(dataDir, fmt.Sprintf("node%d", (i+1)))
		if utils2.Exists(newDir) == false {
			//创建目录
			if err := os.MkdirAll(newDir, 0700); err != nil {
				utils.Fatalf("create output dir [%v]: %v", newDir, err)
			}
		}

		keystoreFile := filepath.Join(newDir, "node.json")
		if utils2.CheckFileExistAndOverride(keystoreFile) == false {
			utils.Logf("Genesis keystore file [%s] failed.", keystoreFile)
			return nil
		}

		utils.Logf("Genesis keystore file [%s]", keystoreFile)

		password := utils2.GetPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true, 0, utils2.MakePasswordList(ctx))

		acct, err := keystore.StoreKeyFile(keystoreFile, password, scryptN, scryptP)

		if err != nil {
			utils.Fatalf("Failed to create account: %v", err)
		}
		fmt.Printf("Address: {%x}\nFilePath: {%s}\n", acct.Address, acct.URL.Path)

		genConfig.Authorities[i] = &entry{acct.Address.String(), fmt.Sprintf("%s%d", authIdPrefix, (31+i))}
	}

	newDir := filepath.Join(dataDir, "approvers")
	if utils2.Exists(newDir) == false {
		//创建目录
		if err := os.MkdirAll(newDir, 0700); err != nil {
			utils.Fatalf("create output dir [%v]: %v", newDir, err)
		}
	}

	utils.Logf("Genesis approvers keystore file")
	for i := 0; i < approverCount; i++ {
		keystoreFile := filepath.Join(newDir, fmt.Sprintf("approver%d.json", (i+1)))
		if utils2.CheckFileExistAndOverride(keystoreFile) == false {
			utils.Logf("Genesis keystore file [%s] failed.", keystoreFile)
			return nil
		}

		utils.Logf("Genesis keystore file [%s]", keystoreFile)

		password := utils2.GetPassPhrase("Your new account is locked with a password. Please give a password. Do not forget this password.", true, 0, utils2.MakePasswordList(ctx))

		acct, err := keystore.StoreKeyFile(keystoreFile, password, scryptN, scryptP)

		if err != nil {
			utils.Fatalf("Failed to create account: %v", err)
		}
		fmt.Printf("Address: {%x}\nFilePath: {%s}\n", acct.Address, acct.URL.Path)

		genConfig.Approvers[i] = &entry{acct.Address.String(), fmt.Sprintf("%s%d", approverIdPrefix, (31+i))}
	}

	jsonStr, err := json.MarshalIndent(genConfig, "", "\t")
	if err != nil {
		utils.Fatalf("%v", err)
	}

	for i := 0; i < authCount; i++ {
		newDir := filepath.Join(dataDir, fmt.Sprintf("node%d", (i+1)))
		genConfigPath := filepath.Join(newDir, "genesis_cfg.json")

		if utils2.CheckFileExistAndOverride(genConfigPath) == false {
			utils.Logf("Genesis config file [%s] failed.", genConfigPath)
			return nil
		}

		if err := ioutil.WriteFile(genConfigPath, jsonStr, 0600); err != nil {
			panic(err)
		}
		utils.Logf("generate %s\n", genConfigPath)
	}

	return nil
}
