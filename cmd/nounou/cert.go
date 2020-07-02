// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/HiNounou029/nounouchain/cmd/nounou/node"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"strings"

	. "github.com/HiNounou029/nounouchain/caclient/aescrypto"
	. "github.com/HiNounou029/nounouchain/caclient/cmd"
	"github.com/HiNounou029/nounouchain/caclient/lib"
	calib "github.com/HiNounou029/nounouchain/caclient/lib"
	. "github.com/HiNounou029/nounouchain/caclient/lib/tcert"
	"github.com/HiNounou029/nounouchain/crypto/ecies"
	"github.com/HiNounou029/nounouchain/crypto/randentropy"
)

func certEnroll(ctx *cli.Context) error {
	var cliCmd = NewCommand("enroll")
	//
	var flagsMap = make(map[string]string)
	flagsMap["u"] = ctx.String(certFlag.Name)
	flagsMap["M"] = ctx.String(certPathFlag.Name)
	cliCmd.Init(flagsMap)

	err := cliCmd.ConfigInit()
	if err != nil {
		return err
	}

	cfgFileName := cliCmd.GetCfgFileName()
	cfg := cliCmd.GetClientCfg()
	resp, err := cfg.Enroll(cfg.URL, filepath.Dir(cfgFileName))
	if err != nil {
		return err
	}

	ID := resp.Identity

	cfgFile, err := ioutil.ReadFile(cfgFileName)
	if err != nil {
		return errors.Wrapf(err, "Failed to read file at '%s'", cfgFileName)
	}

	cfgStr := strings.Replace(string(cfgFile), "<<<ENROLLMENT_ID>>>", ID.GetName(), 1)

	err = ioutil.WriteFile(cfgFileName, []byte(cfgStr), 0644)
	if err != nil {
		return errors.Wrapf(err, "Failed to write file at '%s'", cfgFileName)
	}

	err = ID.Store()
	if err != nil {
		return errors.WithMessage(err, "Failed to store enrollment information")
	}

	// Store root ca
	err = StoreCAChain(cfg, &resp.CAInfo)
	if err != nil {
		return err
	}

	if flagsMap["M"] != "" {
		// Encrypt the private key
		err1 := EncryptPrivateKey(ctx, filepath.Dir(cfgFileName) + "/" + flagsMap["M"] + "/keystore")
		if err1 != nil {
			println("Encrypt the private key failed !")
		}
	}

	return nil
}

func certRegister(ctx *cli.Context) error {
	var cliCmd = NewCommand("register")
	//
	var flagsMap = make(map[string]string)
	flagsMap["id.name"] = ctx.String(idNameFlag.Name)
	flagsMap["id.type"] = "peer"//ctx.String(idTypeFlag.Name)
	flagsMap["id.affiliation"] = "org1.department1"//ctx.String(idAffiliationFlag.Name)
	flagsMap["id.attrs"] = ""//ctx.String(idAttrFlag.Name)
	cliCmd.Init(flagsMap)

	err := cliCmd.ConfigInit()
	if err != nil {
		return err
	}

	client := lib.Client{
		HomeDir: filepath.Dir(cliCmd.GetCfgFileName()),
		Config:  cliCmd.GetClientCfg(),
	}

	id, err := client.LoadMyIdentity()
	if err != nil {
		return err
	}

	//
	cliCmd.GetClientCfg().ID.CAName = cliCmd.GetClientCfg().CAName
	resp, err := id.Register(&cliCmd.GetClientCfg().ID)
	if err != nil {
		return err
	}

	fmt.Printf("Password: %s\n", resp.Secret)

	return nil
}

// Encrypt the private key
func EncryptPrivateKey(ctx *cli.Context, priKeyPath string) error {
	files, _ := ioutil.ReadDir(priKeyPath)
	if len(files) == 1 {
		fileName := files[0].Name()
		filePath := priKeyPath + "/" + fileName
		priKey, err := ioutil.ReadFile(filePath)
		if err != nil{
			return err
		}

		password, _ := getPassword("certificate account")
		println(password)

		// Cryptographic certificates apply for accounts and passwords
		encryptJson, err := EncryptString(string(priKey), password)
		//encryptJson, err := keystore.EncryptString(string(priKey), password, keystore.StandardScryptN, keystore.StandardScryptP)
		if err != nil {
			return err
		}

		ioutil.WriteFile(filePath, encryptJson, 0644)
	} else {
		return errors.New("There are too many private keys !")
	}

	return nil
}

func selfValideNodeCert(ctx *cli.Context, certBuf []byte) error {
	if certBuf == nil {
		return errors.New("Can`t find cert info !")
	}
	peerCert, err := GetCertificate(certBuf)
	if err != nil {
		return  err
	}

	dir := defaultConfigDir()
	caCfg := &calib.CAConfig{}
	caCfg.CA.Chainfile = dir + "/" + ctx.String(certPathFlag.Name) + "/cacerts/rootca.pem"

	// Verify that the certificate is expired
	err = calib.ValidateDates(peerCert)
	if err != nil {
		return err
	}

	//
	var CAChain calib.CA
	CAChain.Config = caCfg
	err = CAChain.VerifyCertificate(peerCert)
	if err != nil {
		return err
	}

	return nil
}

func valCertByConNodeReq(ctx *cli.Context) error{
	dir := defaultConfigDir()
	var certPath = dir + "/" + ctx.String(certPathFlag.Name) + "/signcerts/cert.pem"
	certBuf, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}
	certInfo := make(map[string]interface{})
	certInfo["cert"] = string(certBuf[:])
	bytesData, err := json.Marshal(certInfo)
	if err != nil {
		return err
	}
	reader := bytes.NewReader(bytesData)

	valNodeAddr := ctx.String(remoteNodeAddrFlag.Name)
	if valNodeAddr == ""{
		return errors.New("val node addr is nil !")
	}
	url := "http://" + valNodeAddr + "/verify/cert"
	request, err := http.NewRequest("POST", url, reader)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")
	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return err
	}
	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	m := make(map[string]interface{})
	json.Unmarshal(respBytes, &m)

	if m["result"] == "invalid"{
		return errors.New("invalid node cert !")
	}

	return nil
}

func encryptCert(ctx *cli.Context) error{
	master := loadNodeMaster(ctx)

	dir := defaultConfigDir()
	certPath := dir + "/" + ctx.String(certPathFlag.Name) + "/signcerts/cert.pem"
	certBuf, err := ioutil.ReadFile(certPath)
	if err != nil {
		return errors.New("Read cert file error !")
	}
	//var pubKey PublicKey
	var pubkey ecies.PublicKey
	pubkey.X = master.PrivateKey.X
	pubkey.Y = master.PrivateKey.Y
	pubkey.Curve = master.PrivateKey.Curve
	pubkey.Params = ecies.ECIES_AES128_SHA256

	var rand io.Reader
	rand = randentropy.Reader
	encrpCert, err := ecies.Encrypt(rand, &pubkey, certBuf, nil,nil)
	if err != nil {
		return errors.New("Encrypt cert file error !")
	}

	encCertPath := dir + "/" + ctx.String(certPathFlag.Name) + "/signcerts/encycert.pem"
	ioutil.WriteFile(encCertPath, encrpCert, 0644)

	return nil
}

func decryptCert(ctx *cli.Context) error{
	master := loadNodeMaster(ctx)

	_, err := valEncryptCert(master, ctx.String(certPathFlag.Name))
	if err != nil {
		return err
	}

	return nil
}

func valEncryptCert(master *node.Master, enCertPath string) ([]byte, error){
	dir := defaultConfigDir()
	encCertPath := dir + "/" + enCertPath + "/signcerts/encycert.pem"
	encCertBuf, err := ioutil.ReadFile(encCertPath)
	if err != nil {
		return nil, errors.New("Read encrypt cert file error !")
	}

	var prikey ecies.PrivateKey
	prikey.X = master.PrivateKey.X
	prikey.Y = master.PrivateKey.Y
	prikey.Curve = master.PrivateKey.Curve
	prikey.Params = ecies.ECIES_AES128_SHA256
	prikey.D = master.PrivateKey.D

	orgCert, err := prikey.Decrypt(encCertBuf, nil, nil)
	if err != nil {
		return nil, errors.New("Decrypt encrypt cert file error !")
	}

	return orgCert, nil
}