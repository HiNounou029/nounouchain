// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package logdb_test

import (
	"context"
	"math/big"
	"os"
	"os/user"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/stretchr/testify/assert"
)

func TestEvents(t *testing.T) {
	db, err := logdb.NewMem()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	txEvent := &tx.Event{
		Address: polo.BytesToAddress([]byte("addr")),
		Topics:  []polo.Bytes32{polo.BytesToBytes32([]byte("topic0")), polo.BytesToBytes32([]byte("topic1"))},
		Data:    []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 97, 48},
	}

	header := new(block.Builder).Build().Header()

	for i := 0; i < 100; i++ {
		if err := db.Prepare(header).ForTransaction(polo.BytesToBytes32([]byte("txID")), polo.BytesToAddress([]byte("txOrigin"))).
			Insert(tx.Events{txEvent}, nil).Commit(); err != nil {
			t.Fatal(err)
		}

		header = new(block.Builder).ParentID(header.ID()).Build().Header()
	}

	limit := 5
	t0 := polo.BytesToBytes32([]byte("topic0"))
	t1 := polo.BytesToBytes32([]byte("topic1"))
	addr := polo.BytesToAddress([]byte("addr"))
	es, err := db.FilterEvents(context.Background(), &logdb.EventFilter{
		Range: &logdb.Range{
			Unit: logdb.Block,
			From: 0,
			To:   10,
		},
		Options: &logdb.Options{
			Offset: 0,
			Limit:  uint64(limit),
		},
		Order: logdb.DESC,
		CriteriaSet: []*logdb.EventCriteria{
			&logdb.EventCriteria{
				Address: &addr,
				Topics: [5]*polo.Bytes32{nil,
					nil,
					nil,
					nil,
					nil},
			},
			&logdb.EventCriteria{
				Address: &addr,
				Topics: [5]*polo.Bytes32{&t0,
					&t1,
					nil,
					nil,
					nil},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(es), limit, "limit should be equal")
}

func TestTransfers(t *testing.T) {
	db, err := logdb.NewMem()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

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
			Limit:  uint64(count),
		},
		Order: logdb.DESC,
	}
	ts, err := db.FilterTransfers(context.Background(), tf)
	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(ts), count, "transfers searched")
}

func home() (string, error) {
	// try to get HOME env
	if home := os.Getenv("HOME"); home != "" {
		return home, nil
	}

	//
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	if user.HomeDir != "" {
		return user.HomeDir, nil
	}

	return os.Getwd()
}

func BenchmarkLog(b *testing.B) {
	path, err := home()
	if err != nil {
		b.Fatal(err)
	}

	db, err := logdb.New(path + "/log.db")
	if err != nil {
		b.Fatal(err)
	}
	l := &tx.Event{
		Address: polo.BytesToAddress([]byte("addr")),
		Topics:  []polo.Bytes32{polo.BytesToBytes32([]byte("topic0")), polo.BytesToBytes32([]byte("topic1"))},
		Data:    []byte("data"),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		header := new(block.Builder).Build().Header()
		batch := db.Prepare(header)
		txBatch := batch.ForTransaction(polo.BytesToBytes32([]byte("txID")), polo.BytesToAddress([]byte("txOrigin")))
		for j := 0; j < 100; j++ {
			txBatch.Insert(tx.Events{l}, nil)
			header = new(block.Builder).ParentID(header.ID()).Build().Header()
		}

		if err := batch.Commit(); err != nil {
			b.Fatal(err)
		}
	}
}
