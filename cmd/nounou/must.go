// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package main

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/HiNounou029/nounouchain/common/co"
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/HiNounou029/nounouchain/network"
	"github.com/HiNounou029/nounouchain/network/comm"
	//	"github.com/cloudflare/cfssl/errors"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/ethereum/go-ethereum/rlp"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"errors"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/HiNounou029/nounouchain/cmd/nounou/node"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/fdlimit"
	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/inconshreveable/log15"
	"gopkg.in/urfave/cli.v1"
)

func initLogger(ctx *cli.Context) {
	logLevel := ctx.Int(verbosityFlag.Name)
	log15.Root().SetHandler(log15.LvlFilterHandler(log15.Lvl(logLevel), log15.StderrHandler))
	// set go-ethereum log lvl to Warn
	ethLogHandler := ethlog.NewGlogHandler(ethlog.StreamHandler(os.Stderr, ethlog.TerminalFormat(true)))
	ethLogHandler.Verbosity(ethlog.LvlWarn)
	ethlog.Root().SetHandler(ethLogHandler)
}

func genesisCfgPath(ctx *cli.Context) string {
	configDir := makeConfigDir(ctx)
	return filepath.Join(configDir, "genesis_cfg.json")
}

func selectGenesis(ctx *cli.Context) *genesis.Genesis {
	return genesis.NewProdnet(genesisCfgPath(ctx))
}

func makeConfigDir(ctx *cli.Context) string {
	configDir := ctx.String(configDirFlag.Name)
	if configDir == "" {
		fatal(fmt.Sprintf("unable to infer default config dir, use -%s to specify", configDirFlag.Name))
	}
	if err := os.MkdirAll(configDir, 0700); err != nil {
		fatal(fmt.Sprintf("create config dir [%v]: %v", configDir, err))
	}
	return configDir
}

func makeDataDir(ctx *cli.Context) string {
	dataDir := ctx.String(dataDirFlag.Name)
	if dataDir == "" {
		fatal(fmt.Sprintf("unable to infer default data dir, use -%s to specify", dataDirFlag.Name))
	}
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		fatal(fmt.Sprintf("create data dir [%v]: %v", dataDir, err))
	}
	return dataDir
}

func makeInstanceDir(ctx *cli.Context, gene *genesis.Genesis) string {
	dataDir := makeDataDir(ctx)

	instanceDir := filepath.Join(dataDir, "datastore", fmt.Sprintf("%x", gene.ID().Bytes()[24:]))
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		fatal(fmt.Sprintf("create data dir [%v]: %v", instanceDir, err))
	}
	return instanceDir
}

func openMainDB(ctx *cli.Context, dataDir string) *storage.LevelDB {
	limit, err := fdlimit.Current()
	if err != nil {
		fatal("failed to get fd limit:", err)
	}
	if limit <= 1024 {
		log.Debug("low fd limit, increase it if possible", "limit", limit)
	}

	fileCache := limit / 2
	if fileCache > 1024 {
		fileCache = 1024
	}

	dir := filepath.Join(dataDir, "ledgerstore")
	db, err := storage.New(dir, storage.Options{
		CacheSize:              128,
		OpenFilesCacheCapacity: fileCache,
	})
	if err != nil {
		fatal(fmt.Sprintf("open chain database [%v]: %v", dir, err))
	}
	return db
}

func openLogDB(ctx *cli.Context, dataDir string) *logdb.LogDB {
	dir := filepath.Join(dataDir, "logstore")
	db, err := logdb.New(dir)
	if err != nil {
		fatal(fmt.Sprintf("open log database [%v]: %v", dir, err))
	}
	return db
}

func initChain(gene *genesis.Genesis, mainDB *storage.LevelDB, logDB *logdb.LogDB) *chain.Chain {
	genesisBlock, genesisEvents, err := gene.Build(state.NewCreator(mainDB))
	if err != nil {
		fatal("build genesis block: ", err)
	}

	log.Info("genesis block: ", genesisBlock.String())


	chain, err := chain.New(mainDB, genesisBlock)
	if err != nil {
		fatal("initialize block chain:", err)
	}

	if err := logDB.Prepare(genesisBlock.Header()).
		ForTransaction(polo.Bytes32{}, polo.Address{}).
		Insert(genesisEvents, nil, false).Commit(); err != nil {
		fatal("write genesis events: ", err)
	}
	return chain
}

func masterKeyPath(ctx *cli.Context) string {
	configDir := makeConfigDir(ctx)
	return filepath.Join(configDir, "node.key")
}

func masterKeyStorePath(ctx *cli.Context) string {
	configDir := makeConfigDir(ctx)
	return filepath.Join(configDir, "node.json")
}

func configPath() string {
	path := "conf.json"
	_, err := os.Stat(path)
	if err != nil {
		return filepath.Join("./data/", path)
	}
	return path
}

func beneficiary(ctx *cli.Context) *polo.Address {
	value := ctx.String(beneficiaryFlag.Name)
	if value == "" {
		return nil
	}
	addr, err := polo.ParseAddress(value)
	if err != nil {
		fatal("invalid beneficiary:", err)
	}
	return &addr
}

func readConfig() error {
	file, _ := os.Open(configPath())
	if file == nil {
		return errors.New("conf.json file not found!")
	} else {
		defer file.Close()

		decoder := json.NewDecoder(file)
		err := decoder.Decode(&polo.Conf)
		if err != nil {
			fmt.Println("Error:", err)
			return err
		}
		hw := polo.NewBlake2b()
		rawString := fmt.Sprintf("%v",
			polo.Conf.BlockInterval)
		hw.Write([]byte(rawString))
		var hash polo.Bytes32
		hw.Sum(hash[:0])
		copy(polo.ConfigId[:], hash[:])

	}
	return nil
}

func loadNodeMaster(ctx *cli.Context) *node.Master {
	path := masterKeyStorePath(ctx)
	var err error

	var key *ecdsa.PrivateKey
	_, err = os.Stat(path)
	if err != nil { // file is not exist
		if os.IsNotExist(err) {
			key, err = genPrivateKeyInKeyStore(masterKeyStorePath(ctx), ctx)
			if err != nil {
				fatal("gen node key store error", err)
			} else {
				fatal("node key store error", err)
			}
		}
	} else {
		if key, err = loadPrivateKeyInKeyStore(masterKeyStorePath(ctx), ctx); err != nil {
			fatal("load private key from key store failed. err:\n", err)
		}
	}

	/*
		var err error
		var key *ecdsa.PrivateKey
		if key, err = loadPrivateKeyInKeyStore(masterKeyStorePath(ctx)); err != nil {
			fmt.Printf("load private key from key store failed. err:\r\n %v\r\n", err)
			result, err := readConfirmFromNewTTY("Confirm generate new key even the existed key be overwritten?[Yes|No]:")
			if err != nil || result != "Yes" {
				fatal("confirm failed with result and error:", result, err)
			}
			key, err = genPrivateKeyInKeyStore(masterKeyStorePath(ctx))
			if err != nil {
				fatal("gen node key store error %v", err)
			}
		}
	*/
	master := &node.Master{PrivateKey: key}
	master.Beneficiary = beneficiary(ctx)
	return master
}

type p2pComm struct {
	comm           *comm.Communicator
	p2pSrv         *network.Server
	peersCachePath string
}

func newP2PComm(ctx *cli.Context, chain *chain.Chain, txPool *txpool.TxPool, instanceDir string, rootCaPaht string, isStartWithCert bool, certBuf []byte) *p2pComm {
	configDir := makeConfigDir(ctx)
	key, err := loadOrGeneratePrivateKey(filepath.Join(configDir, "peer.key"))
	if err != nil {
		fatal("load or generate P2P key:", err)
	}
	nat, err := nat.Parse(ctx.String(natFlag.Name))
	if err != nil {
		cli.ShowAppHelp(ctx)
		fmt.Println("parse -nat flag:", err)
		os.Exit(1)
	}
	//如果指定了seed flag，则覆盖默认的bootstrapNodes
	seed := ctx.String(seedFlag.Name)
	if len(seed) > 0 {
		bootstrapNodes = []*discover.Node{
			discover.MustParseNode(seed),
		}
	}

	opts := &network.Options{
		Name:           common.MakeName("polo", fullVersion()),
		PrivateKey:     key,
		MaxPeers:       ctx.Int(maxPeersFlag.Name),
		ListenAddr:     fmt.Sprintf(":%v", ctx.Int(p2pPortFlag.Name)),
		BootstrapNodes: bootstrapNodes,
		NAT:            nat,
	}

	peersCachePath := filepath.Join(instanceDir, "peers.cache")

	if data, err := ioutil.ReadFile(peersCachePath); err != nil {
		if !os.IsNotExist(err) {
			log.Warn("failed to load peers cache", "err", err)
		}
	} else if err := rlp.DecodeBytes(data, &opts.KnownNodes); err != nil {
		log.Warn("failed to load peers cache", "err", err)
	}

	return &p2pComm{
		comm:           comm.New(chain, txPool, rootCaPaht, isStartWithCert, certBuf),
		p2pSrv:         network.New(opts),
		peersCachePath: peersCachePath,
	}
}

func (p *p2pComm) Start() {
	log.Info("starting P2P networking")
	if err := p.p2pSrv.Start(p.comm.Protocols()); err != nil {
		fatal("start P2P server:", err)
	}
	p.comm.Start(p.p2pSrv.NodeInfo())
}

func (p *p2pComm) Stop() {
	log.Info("stopping communicator...")
	p.comm.Stop()

	log.Info("stopping P2P server...")
	p.p2pSrv.Stop()

	log.Info("saving peers cache...")
	nodes := p.p2pSrv.KnownNodes()
	data, err := rlp.EncodeToBytes(nodes)
	if err != nil {
		log.Warn("failed to encode cached peers", "err", err)
		return
	}
	if err := ioutil.WriteFile(p.peersCachePath, data, 0600); err != nil {
		log.Warn("failed to write peers cache", "err", err)
	}
}

func startAPIServer(ctx *cli.Context, handler http.Handler, genesisID polo.Bytes32) (string, func()) {
	addr := ctx.String(apiAddrFlag.Name)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		fatal(fmt.Sprintf("listen API addr [%v]: %v", addr, err))
	}
	timeout := ctx.Int(apiTimeoutFlag.Name)
	if timeout > 0 {
		handler = handleAPITimeout(handler, time.Duration(timeout)*time.Millisecond)
	}
	polo.ApiProtocol = ctx.Int(apiProtocolFlag.Name)
	handler = handleXGenesisID(handler, genesisID)
	handler = handleXPoloChainVersion(handler)
	handler = requestBodyLimit(handler)
	srv := &http.Server{Handler: handler}
	var goes co.Goes
	goes.Go(func() {
		if polo.ApiProtocol == 0 || polo.ApiProtocol == 2 {
			srv.Serve(listener)
		} else {
			err := srv.ServeTLS(listener, "server.crt", "server.key")
			fatal(fmt.Sprintf("ServeTLS addr [%v]: %v", addr, err))
		}
		//		srv.Serve(listener)
	})
	return "http://" + listener.Addr().String() + "/", func() {
		srv.Close()
		goes.Wait()
	}
}

func printStartupMessage(
	chain *chain.Chain,
	master *node.Master,
) {
	bestBlock := chain.BestBlock()
	fmt.Printf(`PoloChain Node Starting...
		Last block   [ %v #%v %v ]
		MinerAddr    [ %v ]
		`,
		bestBlock.Header().ID(), bestBlock.Header().Number(), bestBlock.Header().Timestamp(),
		master.Address())
	fmt.Printf("\r\n")
}
