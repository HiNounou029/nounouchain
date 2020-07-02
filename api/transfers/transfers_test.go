// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package transfers_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HiNounou029/nounouchain/api/transfers"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var ts *httptest.Server

func TestTransfers(t *testing.T) {
	initLogServer(t)
	defer ts.Close()
	getTransfers(t)
}

func getTransfers(t *testing.T) {
	limit := 5
	from := polo.BytesToAddress([]byte("from"))
	to := polo.BytesToAddress([]byte("to"))
	tf := &logdb.TransferFilter{
		CriteriaSet: []*logdb.TransferCriteria{
			&logdb.TransferCriteria{
				TxOrigin:  &from,
				Recipient: &to,
			},
		},
		Range: &logdb.Range{
			Unit: logdb.Block,
			From: 0,
			To:   1000,
		},
		Options: &logdb.Options{
			Offset: 0,
			Limit:  uint64(limit),
		},
		Order: logdb.DESC,
	}
	res := httpPost(t, ts.URL+"/logs/transfer", tf)
	var tLogs []*transfers.FilteredTransfer
	if err := json.Unmarshal(res, &tLogs); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, limit, len(tLogs), "should be `limit` transfers")
}

func initLogServer(t *testing.T) {
	db, err := logdb.NewMem()
	if err != nil {
		t.Fatal(err)
	}

	from := polo.BytesToAddress([]byte("from"))
	to := polo.BytesToAddress([]byte("to"))
	value := big.NewInt(10)
	header := new(block.Builder).Build().Header()
	count := 100
	for i := 0; i < count; i++ {
		transLog := &tx.Transfer{
			Sender:    from,
			Recipient: to,
			Amount:    value,
		}
		header = new(block.Builder).ParentID(header.ID()).Build().Header()
		if err := db.Prepare(header).ForTransaction(polo.Bytes32{}, from).Insert(nil, tx.Transfers{transLog}).
			Commit(); err != nil {
			t.Fatal(err)
		}
	}

	router := mux.NewRouter()
	transfers.New(db).Mount(router, "/logs/transfer")
	ts = httptest.NewServer(router)
}

func httpPost(t *testing.T, url string, obj interface{}) []byte {
	data, err := json.Marshal(obj)
	if err != nil {
		t.Fatal(err)
	}
	res, err := http.Post(url, "application/x-www-form-urlencoded", bytes.NewReader(data))
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
