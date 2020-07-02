// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package main

import (
	"fmt"
	"github.com/HiNounou029/nounouchain/poloclient"
	"github.com/HiNounou029/nounouchain/cmd/bridge"
	"gopkg.in/urfave/cli.v1"
	"math/big"
	"os"
)

var (
	srcChainFlag = cli.IntFlag{
		Name:  "src_chain",
		Value: int(1000),
		Usage: "source chain id",
	}
	dstChainFlag = cli.IntFlag{
		Name:  "dst_chain",
		Value: int(2000),
		Usage: "destination chain id",
	}
	dstAddrFlag = cli.StringFlag{
		Name:  "dst_addr",
		Value: "",
		Usage: "destination addr in destination chain",
	}
	valueFlag = cli.IntFlag{
		Name:  "value",
		Value: int(0),
		Usage: "transfer value from src_addr in src_chain to dst_addr in dst_chain",
	}
	msgFlag = cli.StringFlag{
		Name:  "msg",
		Value: "",
		Usage: "request message from src_addr in src_chain to dst_addr in dst_chain",
	}
	//keystoreFlag = cli.StringFlag{
	//	Name:  "keystore",
	//	Value: "",
	//	Usage: "private key file store of source addr in source chain, src_addr is generated from key",
	//}
	chainConfigFlag = cli.StringFlag{
		Name:  "chain_cfg_path",
		Value: "",
		Usage: "chain configuration path, shared with bridge server",
	}
	flags = []cli.Flag{
		srcChainFlag,
		dstChainFlag,
		dstAddrFlag,
		valueFlag,
		msgFlag,
		//keystoreFlag,
		chainConfigFlag,
	}
)

var chainMap *bridge.ChainMap

func run(ctx *cli.Context) error {
	// load chains
	//var configInfo *bridge.ConfigInfo
	if chainCfgPath := ctx.String(chainConfigFlag.Name); chainCfgPath != "" {
		var err error
		chainMap, err = bridge.InitChains(chainCfgPath, true)
		if err != nil {
			return fmt.Errorf("init chains: %v", err)
		}
	} else {
		return fmt.Errorf("--" + chainConfigFlag.Name + " required")
	}

	//fmt.Printf("%v", configInfo)

	// get src chain
	srcChain := chainMap.ByInt(ctx.Int(srcChainFlag.Name))
	if srcChain == nil {
		return fmt.Errorf("invalid src chain %d", ctx.Int(srcChainFlag.Name))
	}
	//fmt.Printf("%v", srcChain)

	// get dst chain
	dstChain := chainMap.ByInt(ctx.Int(dstChainFlag.Name))
	if dstChain == nil {
		return fmt.Errorf("invalid dst chain %d", ctx.Int(dstChainFlag.Name))
	}
	//fmt.Printf("%v", dstChain)

	// get dst addr
	var dstAddr poloclient.Address
	if dstAddrStr := ctx.String(dstAddrFlag.Name); dstAddrStr != "" {
		dstAddr = poloclient.MustParseAddress(dstAddrStr)
	} else {
		return fmt.Errorf("--" + dstAddrFlag.Name + " required")
	}

	// get value
	value := ctx.Int(valueFlag.Name)

	// get req msg
	msg := ctx.String(msgFlag.Name)

	if value == 0 && msg == "" {
		return fmt.Errorf("value and msg cannot be both empty")
	}
	//fmt.Printf("srcAddr %s\n", srcChain.Bridge.ClientAddr())

	txId, err := srcChain.Bridge.Send(*srcChain.Bridge.ClientAddr(), dstChain.NetworkID, dstAddr,
		big.NewInt(int64(value)), msg)

	if err != nil {
		fmt.Printf("%v", err)
		return err
	}
	fmt.Printf("transaction id %s\n", txId)

	return nil
}

func main() {
	app := cli.App{
		Version:   "1.0",
		Name:      "cross_chain_client",
		Usage:     "send msg/value cross chain",
		Copyright: "2019 PoloChain",
		Flags:     flags,
		Action:    run,
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
