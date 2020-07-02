// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package node_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HiNounou029/nounouchain/api/node"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/HiNounou029/nounouchain/network/comm"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var ts *httptest.Server

func TestNode(t *testing.T) {
	initCommServer(t)
	res := httpGet(t, ts.URL+"/node/network/peers")
	var peersStats map[string]string
	if err := json.Unmarshal(res, &peersStats); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, 0, len(peersStats), "count should be zero")
}

func initCommServer(t *testing.T) {
	db, _ := storage.NewMem()
	stateC := state.NewCreator(db)
	gene := genesis.NewDevnet()

	b, _, err := gene.Build(stateC)
	if err != nil {
		t.Fatal(err)
	}
	chain, _ := chain.New(db, b)
	comm := comm.New(chain, txpool.New(chain, stateC, txpool.Options{
		Limit:           10000,
		LimitPerAccount: 16,
		MaxLifetime:     10 * time.Minute,
		}), "", false, nil)
	router := mux.NewRouter()
	node.New(comm).Mount(router, "/node")
	ts = httptest.NewServer(router)
}

func httpGet(t *testing.T, url string) []byte {
	res, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	r, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	return r
}
