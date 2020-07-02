// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/HiNounou029/nounouchain/api"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/cmd/nounou/node"
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/inconshreveable/log15"
	"github.com/mattn/go-isatty"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
)

var (
	version   string
	gitCommit string
	gitTag    string
	log       = log15.New()

	defaultTxPoolOptions = txpool.Options{
		Limit:           1000000,
		LimitPerAccount: 1000000,
		MaxLifetime:     20 * time.Minute,
	}
)

func fullVersion() string {
	return "release-1.0-1.0"
	versionMeta := "release"
	if gitTag == "" {
		versionMeta = "dev"
	}
	return fmt.Sprintf("%s-%s-%s", version, gitCommit, versionMeta)
}

func main() {
	app := cli.App{
		Version:   fullVersion(),
		Name:      "Polo",
		Usage:     "Polo Chain Node",
		Copyright: "2019 Polo Chain",
		Flags: []cli.Flag{
			configDirFlag,
			seedFlag,
			dataDirFlag,
			beneficiaryFlag,
			apiAddrFlag,
			apiCorsFlag,
			apiTimeoutFlag,
			apiCallGasLimitFlag,
			apiBacktraceLimitFlag,
			apiProtocolFlag,
			verbosityFlag,
			maxPeersFlag,
			p2pPortFlag,
			natFlag,
			certFlag,
			certPathFlag,
			idNameFlag,
			//idTypeFlag,
			//idAffiliationFlag,
			//idAttrFlag,
			encryptFileFlag,
			needCertFlag,
			needValCertFlag,
			remoteNodeAddrFlag,
			accountPwdFlag,
		},
		Action: defaultAction,
		Commands: []cli.Command{
			{
				Name:  "master-key",
				Usage: "import and export master key",
				Flags: []cli.Flag{
					configDirFlag,
					importMasterKeyFlag,
					exportMasterKeyFlag,
				},
				Action: masterKeyAction,
			},
			{
				Name:  "certificate",
				Usage: "Certificate application service",
				Subcommands: []cli.Command{
					{
						Name:   "enroll",
						Usage:  "",
						Action: certEnroll,
						Flags: []cli.Flag{
							certFlag,
							certPathFlag,
						},
						Description: ``,
					},
					{
						Name:   "register",
						Usage:  "",
						Action: certRegister,
						Flags: []cli.Flag{
							idNameFlag,
							//idTypeFlag,
							//idAffiliationFlag,
							//idAttrFlag,
						},
						Description: ``,
					},
				},
			},
			{
				Name:  "asscert",
				Usage: "Associate certificates with accounts",
				Subcommands: []cli.Command{
					{
						Name:   "encrypt",
						Usage:  "",
						Action: encryptCert,
						Flags: []cli.Flag{
							certPathFlag,
							configDirFlag,
							accountPwdFlag,
						},
						Description: ``,
					},
					{
						Name:   "decrypt",
						Usage:  "",
						Action: decryptCert,
						Flags: []cli.Flag{
							certPathFlag,
							configDirFlag,
							accountPwdFlag,
						},
						Description: ``,
					},
				},
			},
		},
	}


	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func defaultAction(ctx *cli.Context) error {
	master := loadNodeMaster(ctx)

	// Certificate self-verification
	var certBuf []byte
	if ctx.Bool(needCertFlag.Name) == true {
		if ctx.String(certPathFlag.Name) == ""{
			return errors.New("Please input cert path by -M flag !")
		}

		certBuff, err := valEncryptCert(master, ctx.String(certPathFlag.Name))
		if err != nil {
			return err
		}
		err = selfValideNodeCert(ctx, certBuff)
		if err != nil {
			fmt.Println(err)
			return err
		}

		certBuf = certBuff
	}

	err := readConfig()
	if err != nil{
		return err
	}

	exitSignal := handleExitSignal()

	defer func() { log.Info("exited") }()

	initLogger(ctx)
	gene := selectGenesis(ctx)
	instanceDir := makeInstanceDir(ctx, gene)

	mainDB := openMainDB(ctx, instanceDir)
	defer func() { log.Info("closing main database..."); mainDB.Close() }()

	logDB := openLogDB(ctx, instanceDir)
	defer func() { log.Info("closing log database..."); logDB.Close() }()

	chain := initChain(gene, mainDB, logDB)

	txPool := txpool.New(chain, state.NewCreator(mainDB), defaultTxPoolOptions)
	defer func() { log.Info("closing tx pool..."); txPool.Close() }()

	// Get current node cert info
	dir := defaultConfigDir()
	rootCaPath := dir + "/" + ctx.String(certPathFlag.Name) + "/cacerts/rootca.pem"
	//certPath := dir + "/" + ctx.String(certPathFlag.Name) + "/signcerts/cert.pem"
	//certBuf, _ := ioutil.ReadFile(certPath)
	p2pcom := newP2PComm(ctx, chain, txPool, instanceDir, rootCaPath, ctx.Bool(needCertFlag.Name), certBuf)

	apiHandler, apiCloser := api.New(chain, state.NewCreator(mainDB), txPool, logDB, p2pcom.comm, ctx.String(apiCorsFlag.Name), uint32(ctx.Int(apiBacktraceLimitFlag.Name)), uint64(ctx.Int(apiCallGasLimitFlag.Name)), rootCaPath)
	defer func() { log.Info("closing API..."); apiCloser() }()

	str, srvCloser := startAPIServer(ctx, apiHandler, chain.GenesisBlock().Header().ID())
	log.Info("api server: ", "listener", str)
	defer func() { log.Info("stopping API server..."); srvCloser() }()

	printStartupMessage(chain, master)

	p2pcom.Start()
	defer p2pcom.Stop()

	return node.New(
		master,
		chain,
		state.NewCreator(mainDB),
		logDB,
		txPool,
		filepath.Join(instanceDir, "btxrecord"),
		p2pcom.comm).
		Run(exitSignal)
}

func masterKeyAction(ctx *cli.Context) error {
	hasImportFlag := ctx.Bool(importMasterKeyFlag.Name)
	hasExportFlag := ctx.Bool(exportMasterKeyFlag.Name)
	if hasImportFlag && hasExportFlag {
		return fmt.Errorf("flag %s and %s are exclusive", importMasterKeyFlag.Name, exportMasterKeyFlag.Name)
	}

	if !hasImportFlag && !hasExportFlag {
		return fmt.Errorf("missing flag, either %s or %s", importMasterKeyFlag.Name, exportMasterKeyFlag.Name)
	}

	if hasImportFlag {
		if isatty.IsTerminal(os.Stdin.Fd()) {
			fmt.Println("Input JSON keystore (end with ^d):")
		}
		keyjson, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			return err
		}

		if err := json.Unmarshal(keyjson, &map[string]interface{}{}); err != nil {
			return errors.WithMessage(err, "unmarshal")
		}
		password, err := readPasswordFromNewTTY("Enter passphrase: ")
		if err != nil {
			return err
		}

		key, err := keystore.DecryptKey(keyjson, password)
		if err != nil {
			return errors.WithMessage(err, "decrypt")
		}

		if err := crypto.SaveECDSA(masterKeyPath(ctx), key.PrivateKey); err != nil {
			return err
		}
		fmt.Println("Master key imported:", polo.Address(key.Address))
		return nil
	}

	if hasExportFlag {
		masterKey, err := loadOrGeneratePrivateKey(masterKeyPath(ctx))
		if err != nil {
			return err
		}

		password, err := readPasswordFromNewTTY("Enter passphrase: ")
		if err != nil {
			return err
		}
		if password == "" {
			return errors.New("non-empty passphrase required")
		}
		confirm, err := readPasswordFromNewTTY("Confirm passphrase: ")
		if err != nil {
			return err
		}

		if password != confirm {
			return errors.New("passphrase confirmation mismatch")
		}

		keyjson, err := keystore.EncryptKey(&keystore.Key{
			PrivateKey: masterKey,
			Address:    crypto.PubkeyToAddress(masterKey.PublicKey),
			Id:         uuid.NewRandom()},
			password, keystore.StandardScryptN, keystore.StandardScryptP)
		if err != nil {
			return err
		}
		if isatty.IsTerminal(os.Stdout.Fd()) {
			fmt.Println("=== JSON keystore ===")
		}
		_, err = fmt.Println(string(keyjson))
		return err
	}
	return nil
}
