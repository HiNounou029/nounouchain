package main

import (
	"gopkg.in/urfave/cli.v1"
	"github.com/HiNounou029/polo-sdk-go/cmd/utils"
)

var (
	// General settings
	DataDirFlag = utils.DirectoryFlag{
		Name:  "datadir",
		Usage: "Data directory for the databases and keystore",
		Value: utils.DirectoryString{utils.DefaultDataDir()},
	}

	PasswordFileFlag = cli.StringFlag{
		Name:  "password",
		Usage: "Password file to use for non-interactive password input",
		Value: "",
	}

	AuthCountFlag = cli.IntFlag{
		Name: "authcount",
		Usage: "Authority count of genesis config, range[1,25]",
		Value: 3,
	}

	ApproverCountFlag = cli.IntFlag{
		Name: "approvercount",
		Usage: "Approver count of genesis config, range[1,25]",
		Value: 3,
	}
)


