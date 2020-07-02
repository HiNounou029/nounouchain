// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package runtime_test

import (
	"math/big"
	"reflect"
	"strings"
	"testing"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/genesis"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/storage"
	"github.com/HiNounou029/nounouchain/vm/runtime"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/HiNounou029/polo-sdk-go/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestResolvedTx(t *testing.T) {
	r, err := newTestResolvedTransaction(t)
	if err != nil {
		t.Fatal(err)
	}

	obValue := reflect.ValueOf(r)
	obType := obValue.Type()
	for i := 0; i < obValue.NumMethod(); i++ {
		if strings.HasPrefix(obType.Method(i).Name, "Test") {
			obValue.Method(i).Call(nil)
		}
	}
}

type testResolvedTransaction struct {
	t            *testing.T
	assert       *assert.Assertions
	chain        *chain.Chain
	stateCreator *state.Creator
}

func newTestResolvedTransaction(t *testing.T) (*testResolvedTransaction, error) {
	db, err := storage.NewMem()
	if err != nil {
		return nil, err
	}

	gen := genesis.NewDevnet()

	stateCreator := state.NewCreator(db)
	parent, _, err := gen.Build(stateCreator)
	if err != nil {
		return nil, err
	}

	c, err := chain.New(db, parent)
	if err != nil {
		return nil, err
	}

	return &testResolvedTransaction{
		t:            t,
		assert:       assert.New(t),
		chain:        c,
		stateCreator: stateCreator,
	}, nil
}

func (tr *testResolvedTransaction) currentState() (*state.State, error) {
	return tr.stateCreator.NewState(tr.chain.BestBlock().Header().StateRoot())
}

func (tr *testResolvedTransaction) TestResolveTransaction() {

	txBuild := func() *tx.Builder {
		return txBuilder(tr.chain.Tag())
	}

	_, err := runtime.ResolveTransaction(txBuild().Build())
	tr.assert.Equal(secp256k1.ErrInvalidSignatureLen, err)

	_, err = runtime.ResolveTransaction(txSign(txBuild().Gas(21000 - 1)))
	tr.assert.NotNil(err)

	address := polo.BytesToAddress([]byte("addr"))
	_, err = runtime.ResolveTransaction(txSign(txBuild().Clause(tx.NewClause(&address).WithValue(big.NewInt(-10)).WithData(nil))))
	tr.assert.NotNil(err)

	_, err = runtime.ResolveTransaction(txSign(txBuild().
		Clause(tx.NewClause(&address).WithValue(math.MaxBig256).WithData(nil)).
		Clause(tx.NewClause(&address).WithValue(math.MaxBig256).WithData(nil)),
	))
	tr.assert.NotNil(err)

	_, err = runtime.ResolveTransaction(txSign(txBuild()))
	tr.assert.Nil(err)
}

func (tr *testResolvedTransaction) TestCommonTo() {

	txBuild := func() *tx.Builder {
		return txBuilder(tr.chain.Tag())
	}

	commonTo := func(tx *tx.Transaction, assert func(interface{}, ...interface{}) bool) {
		resolve, err := runtime.ResolveTransaction(tx)
		if err != nil {
			tr.t.Fatal(err)
		}
		to := resolve.CommonTo()
		assert(to)
	}

	commonTo(txSign(txBuild()), tr.assert.Nil)

	commonTo(txSign(txBuild().Clause(tx.NewClause(nil))), tr.assert.Nil)

	commonTo(txSign(txBuild().Clause(clause()).Clause(tx.NewClause(nil))), tr.assert.Nil)

	address := polo.BytesToAddress([]byte("addr1"))
	commonTo(txSign(txBuild().
		Clause(clause()).
		Clause(tx.NewClause(&address)),
	), tr.assert.Nil)

	commonTo(txSign(txBuild().Clause(clause())), tr.assert.NotNil)
}

func (tr *testResolvedTransaction) TestBuyGas() {
	state, err := tr.currentState()
	if err != nil {
		tr.t.Fatal(err)
	}

	txBuild := func() *tx.Builder {
		return txBuilder(tr.chain.Tag())
	}

	buyGas := func(tx *tx.Transaction) polo.Address {
		resolve, err := runtime.ResolveTransaction(tx)
		if err != nil {
			tr.t.Fatal(err)
		}
		_, returnGas, err := resolve.BuyGas(state)
		tr.assert.Nil(err)
		returnGas(100)
		return resolve.Origin
	}

	tr.assert.Equal(
		genesis.DevAccounts()[0].Address,
		buyGas(txSign(txBuild().Clause(clause().WithValue(big.NewInt(100))))),
	)

	tr.assert.Equal(
		genesis.DevAccounts()[0].Address,
		buyGas(txSign(txBuild().Clause(clause().WithValue(big.NewInt(100))))),
	)

	tr.assert.Equal(
		genesis.DevAccounts()[0].Address,
		buyGas(txSign(txBuild().Clause(clause().WithValue(big.NewInt(100))))),
	)
}

func clause() *tx.Clause {
	address := genesis.DevAccounts()[1].Address
	return tx.NewClause(&address).WithData(nil)
}

func txBuilder(tag byte) *tx.Builder {
	return new(tx.Builder).
		Gas(1000000).
		Expiration(100).
		Nonce(1).
		ChainTag(tag)
}

func txSign(builder *tx.Builder) *tx.Transaction {
	transaction := builder.Build()
	sig, _ := crypto.Sign(transaction.SigningHash().Bytes(), genesis.DevAccounts()[0].PrivateKey)
	return transaction.WithSignature(sig)
}
