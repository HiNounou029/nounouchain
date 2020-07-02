// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package main

import (
	"github.com/inconshreveable/log15"
	cli "gopkg.in/urfave/cli.v1"
)

var (
	configDirFlag = cli.StringFlag{
		Name:   "config-dir",
		Value:  defaultConfigDir(),
		Hidden: true,
		Usage:  "directory for user global configurations",
	}
	dataDirFlag = cli.StringFlag{
		Name:  "data-dir",
		Value: defaultDataDir(),
		Usage: "directory for block-chain databases",
	}
	seedFlag = cli.StringFlag{
		Name:  "seed",
		Value: "",
		Usage: "bootstrap seed node",
	}
	beneficiaryFlag = cli.StringFlag{
		Name:  "beneficiary",
		Usage: "address for block rewards",
	}
	apiAddrFlag = cli.StringFlag{
		Name:  "api-addr",
		Value: "localhost:8669",
		Usage: "API service listening address",
	}
	apiCorsFlag = cli.StringFlag{
		Name:  "api-cors",
		Value: "",
		Usage: "comma separated list of domains from which to accept cross origin requests to API",
	}
	apiTimeoutFlag = cli.IntFlag{
		Name:  "api-timeout",
		Value: 10000,
		Usage: "API request timeout value in milliseconds",
	}
	apiCallGasLimitFlag = cli.IntFlag{
		Name:  "api-call-gas-limit",
		Value: 50000000,
		Usage: "limit contract call gas",
	}
	apiBacktraceLimitFlag = cli.IntFlag{
		Name:  "api-backtrace-limit",
		Value: 1000,
		Usage: "limit the distance between 'position' and best block for subscriptions APIs",
	}
	apiProtocolFlag = cli.IntFlag{
		Name:  "api-protocol",
		Value: 0,
		Usage: "API protocol type (0 http, 1 https, 2 ws, 3 wss)",
	}
	verbosityFlag = cli.IntFlag{
		Name:  "verbosity",
		Value: int(log15.LvlInfo),
		Usage: "log verbosity (0-9)",
	}

	maxPeersFlag = cli.IntFlag{
		Name:  "max-peers",
		Usage: "maximum number of P2P network peers (P2P network disabled if set to 0)",
		Value: 25,
	}
	p2pPortFlag = cli.IntFlag{
		Name:  "p2p-port",
		Value: 11235,
		Usage: "P2P network listening port",
	}
	natFlag = cli.StringFlag{
		Name:  "nat",
		Value: "any",
		Usage: "port mapping mechanism (any|none|upnp|pmp|extip:<IP>)",
	}
	onDemandFlag = cli.BoolFlag{
		Name:  "on-demand",
		Usage: "create new block when there is pending transaction",
	}
	persistFlag = cli.BoolFlag{
		Name:  "persist",
		Usage: "blockchain data storage option, if setted data will be saved to disk",
	}
	gasLimitFlag = cli.IntFlag{
		Name:  "gas-limit",
		Value: 10000000,
		Usage: "block gas limit",
	}
	importMasterKeyFlag = cli.BoolFlag{
		Name:  "import",
		Usage: "import master key from keystore",
	}
	exportMasterKeyFlag = cli.BoolFlag{
		Name:  "export",
		Usage: "export master key to keystore",
	}
	certFlag = cli.StringFlag{
		Name:  "u",
		Usage: "cert enroll",
	}
	certPathFlag = cli.StringFlag{
		Name:  "M",
		Usage: "cert path",
	}
	idNameFlag = cli.StringFlag{
		Name:  "id.name",
		Usage: "cert reg name",
	}
	//idTypeFlag = cli.StringFlag{
	//	Name:  "id.type",
	//	Usage: "cert reg type",
	//}
	//idAffiliationFlag = cli.StringFlag{
	//	Name:  "id.affiliation",
	//	Usage: "cert affiliation",
	//}
	//idAttrFlag = cli.StringFlag{
	//	Name:  "id.attrs",
	//	Usage: "cert attrs",
	//}
	encryptFileFlag = cli.StringFlag{
		Name:  "encryptfile",
		Usage: "encrypt file path",
	}
	needCertFlag = cli.BoolFlag{
		Name:  "needcert",
		Usage: "is need a certificate ",
	}
	needValCertFlag = cli.BoolFlag{
		Name:  "valcert",
		Usage: "verify node certificate ",
	}
	remoteNodeAddrFlag = cli.StringFlag{
		Name:  "remnodeaddr",
		Usage: "remote node address",
	}
	accountPwdFlag = cli.StringFlag{
		Name:  "accpwd",
		Usage: "get account password",
	}
	signCertFlag = cli.StringFlag{
		Name:  "signcert",
		Usage: "sign peer cert, Associate node accounts and certificates",
	}
	verifyCertFlag = cli.StringFlag{
		Name:  "vercert",
		Usage: "verify peer cert",
	}
)
