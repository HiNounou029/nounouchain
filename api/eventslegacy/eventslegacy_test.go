// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package eventslegacy_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/HiNounou029/nounouchain/api/events"
	"github.com/HiNounou029/nounouchain/api/eventslegacy"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

var contractAddr = polo.BytesToAddress([]byte("contract"))
var ts *httptest.Server

func TestEvents(t *testing.T) {
	initEventServer(t)
	defer ts.Close()
	getEvents(t)
}

func getEvents(t *testing.T) {
	t0 := polo.BytesToBytes32([]byte("topic0"))
	t1 := polo.BytesToBytes32([]byte("topic1"))
	limit := 5
	filter := &eventslegacy.FilterLegacy{
		Range: &logdb.Range{
			Unit: "",
			From: 0,
			To:   10,
		},
		Options: &logdb.Options{
			Offset: 0,
			Limit:  uint64(limit),
		},
		Order:   "",
		Address: &contractAddr,
		TopicSets: []*eventslegacy.TopicSet{
			&eventslegacy.TopicSet{
				Topic0: &t0,
			},
			&eventslegacy.TopicSet{
				Topic1: &t1,
			},
		},
	}
	res := httpPost(t, ts.URL+"/logs/events?address="+contractAddr.String(), filter)
	var logs []*events.FilteredEvent
	if err := json.Unmarshal(res, &logs); err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, limit, len(logs), "should be `limit` logs")
}

func initEventServer(t *testing.T) {
	db, err := logdb.NewMem()
	if err != nil {
		t.Fatal(err)
	}
	txEv := &tx.Event{
		Address: contractAddr,
		Topics:  []polo.Bytes32{polo.BytesToBytes32([]byte("topic0")), polo.BytesToBytes32([]byte("topic1"))},
		Data:    []byte("data"),
	}

	header := new(block.Builder).Build().Header()
	for i := 0; i < 100; i++ {
		if err := db.Prepare(header).ForTransaction(polo.BytesToBytes32([]byte("txID")), polo.BytesToAddress([]byte("txOrigin"))).
			Insert(tx.Events{txEv}, nil).Commit(); err != nil {
			if err != nil {
				t.Fatal(err)
			}
		}
		header = new(block.Builder).ParentID(header.ID()).Build().Header()
	}

	router := mux.NewRouter()
	eventslegacy.New(db).Mount(router, "/logs/events")
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
