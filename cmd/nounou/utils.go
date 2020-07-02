// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/pborman/uuid"
	"github.com/pkg/errors"
	"gopkg.in/urfave/cli.v1"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/mattn/go-tty"
	"github.com/HiNounou029/nounouchain/api"
)

func fatal(args ...interface{}) {
	var w io.Writer
	if runtime.GOOS == "windows" {
		// The SameFile check below doesn't work on Windows.
		// stdout is unlikely to get redirected though, so just print there.
		w = os.Stdout
	} else {
		outf, _ := os.Stdout.Stat()
		errf, _ := os.Stderr.Stat()
		if outf != nil && errf != nil && os.SameFile(outf, errf) {
			w = os.Stderr
		} else {
			w = io.MultiWriter(os.Stdout, os.Stderr)
		}
	}
	fmt.Fprint(w, "Fatal: ")
	fmt.Fprintln(w, args...)
	os.Exit(1)
}

func loadOrGeneratePrivateKey(path string) (*ecdsa.PrivateKey, error) {
	key, err := crypto.LoadECDSA(path)
	if err == nil {
		return key, nil
	}

	if !os.IsNotExist(err) {
		return nil, err
	}

	key, err = crypto.GenerateKey()
	if err != nil {
		return nil, err
	}
	if err := crypto.SaveECDSA(path, key); err != nil {
		return nil, err
	}
	return key, nil
}

func loadPrivateKeyInKeyStore(path string, ctx *cli.Context) (*ecdsa.PrivateKey, error) {
	keyjson, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(keyjson, &map[string]interface{}{}); err != nil {
		return nil, errors.WithMessage(err, "unmarshal")
	}

	password := ctx.String(accountPwdFlag.Name)
	if password == "" {
		password, err = readPasswordFromNewTTY("Enter loading passphrase: ")
		if err != nil {
			return nil, err
		}
	}

	key, err := keystore.DecryptKey(keyjson, password)
	if err != nil {
		return nil, errors.WithMessage(err, "decrypt")
	}

	fmt.Printf("path[%s] keystore imported\n", path)
	return key.PrivateKey, nil
}

func genPrivateKeyInKeyStore(path string, ctx *cli.Context) (*ecdsa.PrivateKey, error) {
	masterKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	password := ctx.String(accountPwdFlag.Name)
	if password == ""{
		password, err := readPasswordFromNewTTY("Enter generating passphrase: ")
		if err != nil {
			return nil, err
		}
		if password == "" {
			return nil, errors.New("non-empty passphrase required")
		}
		confirm, err := readPasswordFromNewTTY("Confirm generating passphrase: ")
		if err != nil {
			return nil, err
		}
		if password != confirm {
			return nil, errors.New("generating passphrase confirmation mismatch")
		}
	}

	keyjson, err := keystore.EncryptKey(&keystore.Key{
		PrivateKey: masterKey,
		Address:    crypto.PubkeyToAddress(masterKey.PublicKey),
		Id:         uuid.NewRandom()},
		password, keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return nil, err
	}
	err = ioutil.WriteFile(path, []byte(keyjson), 0600)
	return masterKey, err
}

func defaultConfigDir() string {
	return xb_defaultdir()

	if home := homeDir(); home != "" {
		return filepath.Join(home, ".polochain.com")
	}
	return ""
}

func xb_defaultdir() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		//		log.Fatal(err)
	}
	return dir

}

// copy from go-ethereum
func defaultDataDir() string {
	return xb_defaultdir()

	// Try to place the data folder in the user's home dir
	if home := homeDir(); home != "" {
		if runtime.GOOS == "darwin" {
			return filepath.Join(home, "Library", "Application Support", ".polochain.com")
		} else if runtime.GOOS == "windows" {
			return filepath.Join(home, "AppData", "Roaming", ".polochain.com")
		} else {
			return filepath.Join(home, ".polochain.com")
		}
	}
	// As we cannot guess a stable location, return empty and handle later
	return ""
}

func homeDir() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	if usr, err := user.Current(); err == nil {
		return usr.HomeDir
	}
	return ""
}

func handleExitSignal() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		exitSignalCh := make(chan os.Signal)
		signal.Notify(exitSignalCh, os.Interrupt, os.Kill, syscall.SIGTERM)

		select {
		case sig := <-exitSignalCh:
			log.Info("exit signal received", "signal", sig)
			cancel()
		}
	}()
	return ctx
}

// middleware to limit request body size.
func requestBodyLimit(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 96*1000)
		h.ServeHTTP(w, r)
	})
}

// middleware to verify 'x-genesis-id' header in request, and set to response headers.
func handleXGenesisID(h http.Handler, genesisID polo.Bytes32) http.Handler {
	const headerKey = "x-genesis-id"
	expectedID := genesisID.String()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		actualID := r.Header.Get(headerKey)
		w.Header().Set(headerKey, expectedID)
		if actualID != "" && actualID != expectedID {
			io.Copy(ioutil.Discard, r.Body)
			http.Error(w, "genesis id mismatch", http.StatusForbidden)
			return
		}
		h.ServeHTTP(w, r)
	})
}

// middleware to set 'x-polochain-ver' to response headers.
func handleXPoloChainVersion(h http.Handler) http.Handler {
	const headerKey = "x-polochain-ver"
	ver := api.ApiVer
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(headerKey, ver)
		h.ServeHTTP(w, r)
	})
}

// middleware for http request timeout.
func handleAPITimeout(h http.Handler, timeout time.Duration) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), timeout)
		defer cancel()
		r = r.WithContext(ctx)
		h.ServeHTTP(w, r)
	})
}

func readPasswordFromNewTTY(prompt string) (string, error) {
	t, err := tty.Open()
	if err != nil {
		return "", err
	}
	defer t.Close()
	fmt.Fprint(t.Output(), prompt)
	pass, err := t.ReadPasswordNoEcho()
	if err != nil {
		return "", err
	}
	return pass, err
}

func readConfirmFromNewTTY(prompt string) (string, error) {
	t, err := tty.Open()
	if err != nil {
		return "", err
	}
	defer t.Close()
	fmt.Fprint(t.Output(), prompt)
	confirm, err := t.ReadString()
	if err != nil {
		return "", err
	}
	return confirm, err
}

func getPassword(tipStr string) (string, error) {
	var firstPwd, conPwd string
	strTip := "Please enter a password for " + tipStr
	fmt.Print(strTip)
	fmt.Scanln(&firstPwd)
	fmt.Print("Please confirm password : ")
	fmt.Scanln(&conPwd)

	count := 1
	for {
		if firstPwd == conPwd {
			break
		} else if count < 3{
			fmt.Println("Please confirm password : ")
			fmt.Scanln(&conPwd)
			count++
		} else {
			var err error = errors.New("The two passwords do not match")
			return "", err
		}
	}

	return firstPwd, nil
}