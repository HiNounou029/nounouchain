package main

import (
	"gopkg.in/urfave/cli.v1"
	"os"
	"fmt"
	utils2 "github.com/HiNounou029/nounouchain/cmd/utils"
	"github.com/HiNounou029/polo-sdk-go/cmd/utils"
	//"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/polo-sdk-go/client"
	"github.com/HiNounou029/polo-sdk-go"
	"github.com/HiNounou029/polo-sdk-go/common"
	"github.com/HiNounou029/nounouchain/polo"
	"strings"
)

var (
	keystorePathFlag = cli.StringFlag{
		Name:  "keystore",
		Value: "",
		Usage: "the file path of approver keystore file path",
	}

	passwordFlag = cli.StringFlag{
		Name:  "password",
		Value: "",
		Usage: "the passphrase to encrypt keystore file",
	}

	addressFlag = cli.StringFlag{
		Name: "addr",
		Value: "",
		Usage: "address of account",
	}

	idFlag = cli.StringFlag{
		Name: "id",
		Value: "",
		Usage: "id of account",
	}

	flags = []cli.Flag{
		keystorePathFlag,
		passwordFlag,
		addressFlag,
		utils.ChainNodeUrlFlag,
		idFlag,
	}

	proposeCmd = cli.Command{
		Name: "propose",
		Usage: "propose authority node",
		Flags:[]cli.Flag{
			keystorePathFlag,
			passwordFlag,
			addressFlag,
			utils.ChainNodeUrlFlag,
			idFlag,
		},
		Action:utils2.MigrateFlags(proposeAuthority),
	}

	approveCmd = cli.Command{
		Name: "approve",
		Usage: "approve authority node",
		Flags:[]cli.Flag{
			keystorePathFlag,
			passwordFlag,
			utils.ChainNodeUrlFlag,
		},
		Action:utils2.MigrateFlags(approveAuthority),
	}

	executeCmd = cli.Command{
		Name: "execute",
		Usage: "execute",
		Action: utils2.MigrateFlags(executeAuthority),
		Flags: []cli.Flag{
			keystorePathFlag,
			passwordFlag,
			utils.ChainNodeUrlFlag,
		},
	}

)


func main() {
	app := cli.App{
		Version:   "1.0",
		Name:      "genesis",
		Usage:     "generate genesis configuration files",
		Copyright: "2019 PoloChain",
		Flags:     flags,
		Action:    run,
		Commands:  []cli.Command{
			proposeCmd,
			approveCmd,
			executeCmd,
		},
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {

	return nil
}

func proposeAuthority(ctx *cli.Context) error {

	action := ctx.Args().First()
	if len(action) == 0 {
		utils.Fatalf("Action must be given as argument")
	}

	if action != "add" && action != "revoke" {
		utils.Fatalf("Action must be 'add' or 'revoke'.")
	}


	keystoreFile := ctx.GlobalString(keystorePathFlag.Name)
	if len(keystoreFile) == 0 {
		utils.Fatalf("Approver keystore file must be given as argument")
	}

	if utils2.Exists(keystoreFile) == false {
		utils.Fatalf("KeystoreFile: %s not exist.", keystoreFile)
	}

	authAddr := ctx.GlobalString(addressFlag.Name)
	if len(authAddr) == 0 {
		utils.Fatalf("Authority address must be given as argument")
	}

	addr, err := client.ParseAddress(authAddr)
	if err != nil {
		utils.Fatalf("Address: %s incorrect, %v", authAddr, err)
	}

	actionAuth, exist := builtin.Authority.ABI.MethodByName(action)
	if exist == false {
		utils.Fatalf("action %s not exist.", action)
	}

	var authData []byte

	if action == "add" {

		authId := ctx.GlobalString(idFlag.Name)
		if len(authId) == 0 {
			utils.Fatalf("Authority id must be given as argument")
		}

		//function add(address _nodeMaster, address _endorsor, bytes32 _identity)
		authData, err = actionAuth.EncodeInput(addr, addr, common.BytesToBytes32([]byte(authId)))
	} else {
		//function revoke(address _nodeMaster)
		authData, err = actionAuth.EncodeInput(addr)
	}

	if err != nil {
		utils.Fatalf("Encode error: %v", err)
	}

	//function propose(address _target, bytes _data)
	actionPropose, _ := builtin.Executor.ABI.MethodByName("propose")
	proposeData, err := actionPropose.EncodeInput(builtin.Authority.Address, authData)
	if err != nil {
		utils.Fatalf("Encode error: %v", err)
	}


	passphrase := utils2.GetPassPhrase("Please enter the password of the keystore file.", false, 0, utils2.MakePasswordList(ctx))

	nodeUrl := ctx.GlobalString(utils.ChainNodeUrlFlag.Name)
	sdk := polo_go_sdk.NewPoloSdk()
	sdk.SetHost(nodeUrl)

	signer, err := sdk.UnlockAccount(keystoreFile, passphrase, 0)

	if err != nil {
		utils.Fatalf("UnlockAccount error: %v", err)
		return nil
	}

	ch := make(chan int)
	go func() {

		onReceipt := func(receipt *client.Receipt, err error) {

			if err != nil {
				utils.Fatalf("deploy contract receipt failed: %v", err)
			}

			if len(receipt.Outputs) == 0 {
				utils.Fatalf("receipt outputs empty")
			}

			fmt.Printf("block num: %d, block time: %d\n", receipt.Meta.BlockNumber, receipt.Meta.BlockTimestamp)

			//event Proposal(bytes32 indexed proposalID, bytes32 action);
			execAddr := strings.ToLower(builtin.Executor.Address.String())
			for i := 0; i < len(receipt.Outputs); i++ {
				for j := 0; j < len(receipt.Outputs[i].Events); j++ {
					event := receipt.Outputs[i].Events[j]
					evntAddr := strings.ToLower(event.Address.String())
					if evntAddr == execAddr {
						_, ret := builtin.Executor.ABI.EventByID(polo.Bytes32(event.Topics[0]))
						if ret {

							actionBytes := common.FromHex(event.Data)

							if err == nil {
								utils.Logf("proposalID: %s, action: %s", event.Topics[1].String(), common.BytesToNormalString(actionBytes))
							} else {
								utils.Fatalf("Decode err: %v", err)
							}

						}
					}
				}
			}

			ch <- 22
		}

		result, err := sdk.InvokeContract2(proposeData, builtin.Executor.Address.String(),0, signer, onReceipt)
		if err != nil {
			utils.Fatalf("InvokeContract error: %v", err)
			return
		}
		fmt.Printf("invoke contract result: %s\n", string(result))

	}()
	<-ch

	return nil
}

func approveAuthority(ctx *cli.Context) error {
	return exectuorAction(ctx, "approve")
}


func executeAuthority(ctx *cli.Context) error {
	return exectuorAction(ctx, "execute")
}

func exectuorAction(ctx *cli.Context, action string) error {

	proposalID := ctx.Args().First()
	if len(proposalID) == 0 {
		utils.Fatalf("ProposalID must be given as argument")
	}

	bytesProposalId, err := common.ParseBytes32(proposalID)
	if err != nil {
		utils.Fatalf("ParseBytes32, err: %v", err)
	}


	keystoreFile := ctx.GlobalString(keystorePathFlag.Name)
	if len(keystoreFile) == 0 {
		utils.Fatalf("Approver keystore file must be given as argument")
	}

	if utils2.Exists(keystoreFile) == false {
		utils.Fatalf("KeystoreFile: %s not exist.", keystoreFile)
	}

	//function execute(bytes32 _proposalID) or function approve(bytes32 _proposalID)
	actionApprove, exist := builtin.Executor.ABI.MethodByName(action)

	if exist == false {
		utils.Fatalf("action %s not exist.", action)
	}

	approveData, err := actionApprove.EncodeInput(bytesProposalId)
	if err != nil {
		utils.Fatalf("Encode error: %v", err)
	}


	passphrase := utils2.GetPassPhrase("Please enter the password of the keystore file.", false, 0, utils2.MakePasswordList(ctx))

	nodeUrl := ctx.GlobalString(utils.ChainNodeUrlFlag.Name)
	sdk := polo_go_sdk.NewPoloSdk()
	sdk.SetHost(nodeUrl)

	signer, err := sdk.UnlockAccount(keystoreFile, passphrase, 0)

	if err != nil {
		utils.Fatalf("UnlockAccount error: %v", err)
		return nil
	}


	ch := make(chan int)
	go func() {

		onReceipt := func(receipt *client.Receipt, err error) {

			if err != nil {
				utils.Fatalf("deploy contract receipt failed: %v", err)
			}

			if len(receipt.Outputs) == 0 {
				utils.Fatalf("receipt outputs empty")
			}

			fmt.Printf("block num: %d, block time: %d\n", receipt.Meta.BlockNumber, receipt.Meta.BlockTimestamp)

			execAddr := strings.ToLower(builtin.Executor.Address.String())
			authAddr := strings.ToLower(builtin.Authority.Address.String())

			for i := 0; i < len(receipt.Outputs); i++ {
				for j := 0; j < len(receipt.Outputs[i].Events); j++ {
					event := receipt.Outputs[i].Events[j]
					utils.Logf("event addr: %s", event.Address.String())
					evntAddr := strings.ToLower(event.Address.String())
					if evntAddr == execAddr {
						//event Proposal(bytes32 indexed proposalID, bytes32 action);
						_, ret := builtin.Executor.ABI.EventByID(polo.Bytes32(event.Topics[0]))
						if ret {

							actionBytes := common.FromHex(event.Data)
							utils.Logf("proposalID: %s, action: %s", event.Topics[1].String(), common.BytesToNormalString(actionBytes))

						}
					}

					if evntAddr == authAddr {
						//event Candidate(address indexed nodeMaster, bytes32 action);
						addr := polo.BytesToAddress(event.Topics[1].Bytes())
						_, ret := builtin.Authority.ABI.EventByID(polo.Bytes32(event.Topics[0]))
						if ret {
							actionBytes := common.FromHex(event.Data)
							utils.Logf("_nodeMaster: %s, action: %s", addr.String(), common.BytesToNormalString(actionBytes))
						}
					}
				}
			}

			ch <- 22
		}

		result, err := sdk.InvokeContract2(approveData, builtin.Executor.Address.String(),0, signer, onReceipt)
		if err != nil {
			utils.Fatalf("InvokeContract error: %v", err)
			return
		}
		fmt.Printf("invoke contract result: %s\n", string(result))

	}()
	<-ch

	return nil
}



