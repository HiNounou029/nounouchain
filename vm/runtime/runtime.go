// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package runtime

import (
	"math/big"
	"sync/atomic"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	Tx "github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/HiNounou029/nounouchain/vm"
	"github.com/HiNounou029/nounouchain/vm/runtime/statedb"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/params"
	"github.com/pkg/errors"
	"github.com/inconshreveable/log15"
)

var log = log15.New("pkg", "runtime")

var (
	//energyTransferEvent     *abi.Event
	prototypeSetMasterEvent *abi.Event
	nativeCallReturnGas     uint64 = 1562 // see test case for calculation
)

func init() {
	var found bool
	//if energyTransferEvent, found = builtin.Energy.ABI.EventByName("Transfer"); !found {
	//	panic("transfer event not found")
	//}
	if prototypeSetMasterEvent, found = builtin.Prototype.Events().EventByName("$Master"); !found {
		panic("$Master event not found")
	}
}

var chainConfig = params.ChainConfig{
	ChainID:             big.NewInt(0),
	HomesteadBlock:      big.NewInt(0),
	DAOForkBlock:        big.NewInt(0),
	DAOForkSupport:      false,
	EIP150Block:         big.NewInt(0),
	EIP150Hash:          common.Hash{},
	EIP155Block:         big.NewInt(0),
	EIP158Block:         big.NewInt(0),
	ByzantiumBlock:      big.NewInt(0),
	ConstantinopleBlock: nil,
	Ethash:              nil,
	Clique:              nil,
}

// Output output of clause execution.
type Output struct {
	Data            []byte
	Events          tx.Events
	Transfers       tx.Transfers
	LeftOverGas     uint64
	RefundGas       uint64
	VMErr           error         // VMErr identify the execution result of the contract function, not evm function's err.
	ContractAddress *polo.Address // if create a new contract, or is nil.
}

type TransactionExecutor struct {
	HasNextClause func() bool
	NextClause    func() (gasUsed uint64, output *Output, err error)
	Finalize      func() (*tx.Receipt, error)
}

// Runtime bases on EVM and the builtin contract.
type Runtime struct {
	vmConfig   vm.Config
	seeker     *chain.Seeker
	state      *state.State
	ctx        *xenv.BlockContext
	forkConfig polo.ForkConfig
}

// New create a Runtime object.
func New(
	seeker *chain.Seeker,
	state *state.State,
	ctx *xenv.BlockContext,
) *Runtime {
	rt := Runtime{
		seeker: seeker,
		state:  state,
		ctx:    ctx,
	}
	if seeker != nil {
		rt.forkConfig = polo.GetForkConfig(seeker.GenesisID())
	} else {
		// for genesis building stage
		rt.forkConfig = polo.NoFork
	}
	return &rt
}

func (rt *Runtime) Seeker() *chain.Seeker       { return rt.seeker }
func (rt *Runtime) State() *state.State         { return rt.state }
func (rt *Runtime) Context() *xenv.BlockContext { return rt.ctx }

// SetVMConfig config VM.
// Returns this runtime.
func (rt *Runtime) SetVMConfig(config vm.Config) *Runtime {
	rt.vmConfig = config
	return rt
}

func (rt *Runtime) newEVM(stateDB *statedb.StateDB, clauseIndex uint32, txCtx *xenv.TransactionContext) *vm.EVM {
	var lastNonNativeCallGas uint64
	return vm.NewEVM(vm.Context{
		CanTransfer: func(_ vm.StateDB, addr common.Address, amount *big.Int) bool {
			return stateDB.GetBalance(addr).Cmp(amount) >= 0
		},
		Transfer: func(_ vm.StateDB, sender, recipient common.Address, amount *big.Int) {
			if amount.Sign() == 0 {
				return
			}
			stateDB.SubBalance(common.Address(sender), amount)
			stateDB.AddBalance(common.Address(recipient), amount)

			if rt.ctx.Number >= rt.forkConfig.FixTransferLog {
				// `amount` will be recycled by evm(OP_CALL) right after this function return,
				// which leads to incorrect transfer log.
				// Make a copy to prevent it.
				amount = new(big.Int).Set(amount)
			}

			stateDB.AddTransfer(&tx.Transfer{
				Sender:    polo.Address(sender),
				Recipient: polo.Address(recipient),
				Amount:    amount,
			})
		},
		GetHash: func(num uint64) common.Hash {
			return common.Hash(rt.seeker.GetID(uint32(num)))
		},
		NewContractAddress: func(_ *vm.EVM, counter uint32) common.Address {
			return common.Address(polo.CreateContractAddress(txCtx.ID, clauseIndex, counter))
		},
		InterceptContractCall: func(evm *vm.EVM, contract *vm.Contract, readonly bool) ([]byte, error, bool) {
			if evm.Depth() < 2 {
				lastNonNativeCallGas = contract.Gas
				// skip direct calls
				return nil, nil, false
			}

			if contract.Address() != contract.Caller() {
				lastNonNativeCallGas = contract.Gas
				// skip native calls from other contract
				return nil, nil, false
			}

			abi, run, found := builtin.FindNativeCall(polo.Address(contract.Address()), contract.Input)
			if !found {
				lastNonNativeCallGas = contract.Gas
				return nil, nil, false
			}

			if readonly && !abi.Const() {
				panic("invoke non-const method in readonly env")
			}

			if contract.Value().Sign() != 0 {
				// reject value transfer on call
				panic("value transfer not allowed")
			}

			// here we return call gas and extcodeSize gas for native calls, to make
			// builtin contract cheap.
			contract.Gas += nativeCallReturnGas
			if contract.Gas > lastNonNativeCallGas {
				panic("serious bug: native call returned gas over consumed")
			}

			ret, err := xenv.New(abi, rt.seeker, rt.state, rt.ctx, txCtx, evm, contract).Call(run)
			return ret, err, true
		},
		OnCreateContract: func(_ *vm.EVM, contractAddr, caller common.Address) {
			// set master for created contract
			rt.state.SetMaster(polo.Address(contractAddr), polo.Address(caller))

			data, err := prototypeSetMasterEvent.Encode(caller)
			if err != nil {
				panic(err)
			}

			stateDB.AddLog(&types.Log{
				Address: common.Address(contractAddr),
				Topics:  []common.Hash{common.Hash(prototypeSetMasterEvent.ID())},
				Data:    data,
			})
		},
		OnSuicideContract: func(_ *vm.EVM, contractAddr, tokenReceiver common.Address) {
			if amount := stateDB.GetBalance(contractAddr); amount.Sign() != 0 {
				stateDB.AddBalance(tokenReceiver, amount)
				stateDB.SubBalance(contractAddr, amount)

				stateDB.AddTransfer(&tx.Transfer{
					Sender:    polo.Address(contractAddr),
					Recipient: polo.Address(tokenReceiver),
					Amount:    amount,
				})
			}
		},
		Origin:      common.Address(txCtx.Origin),
		GasPrice:    txCtx.GasPrice,
		Coinbase:    common.Address(rt.ctx.Beneficiary),
		GasLimit:    rt.ctx.GasLimit,
		BlockNumber: new(big.Int).SetUint64(uint64(rt.ctx.Number)),
		Time:        new(big.Int).SetUint64(rt.ctx.Time),
		Difficulty:  &big.Int{},
	}, stateDB, &chainConfig, rt.vmConfig)
}

// ExecuteClause executes single clause.
func (rt *Runtime) ExecuteClause(
	clause *tx.Clause,
	clauseIndex uint32,
	gas uint64,
	txCtx *xenv.TransactionContext,
) *Output {
	exec, _ := rt.PrepareClause(clause, clauseIndex, gas, txCtx)
	output, _ := exec()
	return output
}

// PrepareClause prepare to execute clause.
// It allows to interrupt execution.
func (rt *Runtime) PrepareClause(
	clause *tx.Clause,
	clauseIndex uint32,
	gas uint64,
	txCtx *xenv.TransactionContext,
) (exec func() (output *Output, interrupted bool), interrupt func()) {
	var (
		stateDB       = statedb.New(rt.state)
		evm           = rt.newEVM(stateDB, clauseIndex, txCtx)
		data          []byte
		leftOverGas   uint64
		vmErr         error
		contractAddr  *polo.Address
		interruptFlag uint32
	)

	exec = func() (*Output, bool) {
		if clause.To() == nil {
			var caddr common.Address
			data, caddr, leftOverGas, vmErr = evm.Create(vm.AccountRef(txCtx.Origin), clause.Data(), gas, clause.Value())
			contractAddr = (*polo.Address)(&caddr)
		} else {
			data, leftOverGas, vmErr = evm.Call(vm.AccountRef(txCtx.Origin), common.Address(*clause.To()), clause.Data(), gas, clause.Value())
		}

		interrupted := atomic.LoadUint32(&interruptFlag) != 0
		output := &Output{
			Data:            data,
			LeftOverGas:     leftOverGas,
			RefundGas:       stateDB.GetRefund(),
			VMErr:           vmErr,
			ContractAddress: contractAddr,
		}
		output.Events, output.Transfers = stateDB.GetLogs()
		return output, interrupted
	}

	interrupt = func() {
		atomic.StoreUint32(&interruptFlag, 1)
		evm.Cancel()
	}
	return
}

// ExecuteTransaction executes a transaction.
// If some clause failed, receipt.Outputs will be nil and vmOutputs may shorter than clause count.
func (rt *Runtime) ExecuteTransaction(tx *tx.Transaction) (receipt *tx.Receipt, err error) {
	executor, err := rt.PrepareTransaction(tx)
	if err != nil {
		return nil, err
	}
	for executor.HasNextClause() {
		if _, _, err := executor.NextClause(); err != nil {
			return nil, err
		}
	}
	return executor.Finalize()
}

// PrepareTransaction prepare to execute tx.
func (rt *Runtime) PrepareTransaction(tx *tx.Transaction) (*TransactionExecutor, error) {
	resolvedTx, err := ResolveTransaction(tx)
	if err != nil {
		return nil, err
	}

	gasPrice, returnGas, err := resolvedTx.BuyGas(rt.state)
	if err != nil {
		return nil, err
	}

	// ResolveTransaction has checked that tx.Gas() >= IntrinsicGas
	leftOverGas := tx.Gas() - resolvedTx.IntrinsicGas
	// checkpoint to be reverted when clause failure.
	checkpoint := rt.state.NewCheckpoint()

	txCtx := resolvedTx.ToContext(gasPrice, rt.ctx.Number, rt.seeker.GetID)

	txOutputs := make([]*Tx.Output, 0, len(resolvedTx.Clauses))
	reverted := false
	finalized := false

	hasNext := func() bool {
		return !reverted && len(txOutputs) < len(resolvedTx.Clauses)
	}

	return &TransactionExecutor{
		HasNextClause: hasNext,
		NextClause: func() (gasUsed uint64, output *Output, err error) {
			if !hasNext() {
				return 0, nil, errors.New("no more clause")
			}
			nextClauseIndex := uint32(len(txOutputs))
			output = rt.ExecuteClause(resolvedTx.Clauses[nextClauseIndex], nextClauseIndex, leftOverGas, txCtx)
			//if len(output.Data) > 0 {
			//	fmt.Println("output: " + common.Bytes2Hex(output.Data))
			//}
			gasUsed = leftOverGas - output.LeftOverGas
			leftOverGas = output.LeftOverGas

			// Apply refund counter, capped to half of the used gas.
			refund := gasUsed / 2
			if refund > output.RefundGas {
				refund = output.RefundGas
			}

			// won't overflow
			leftOverGas += refund

			if output.VMErr != nil {
				// vm exception here
				// revert all executed clauses
				rt.state.RevertTo(checkpoint)
				reverted = true
				txOutputs = nil
				log.Error("VMErr", "err", output.VMErr, "txId", tx.ID())
				return
			}
			txOutputs = append(txOutputs, &Tx.Output{Events: output.Events, Transfers: output.Transfers})
			return
		},
		Finalize: func() (*Tx.Receipt, error) {
			if hasNext() {
				return nil, errors.New("not all clauses processed")
			}
			if finalized {
				return nil, errors.New("already finalized")
			}
			finalized = true

			receipt := &Tx.Receipt{
				Reverted: reverted,
				Outputs:  txOutputs,
				GasUsed:  tx.Gas() - leftOverGas,
				GasPayer: resolvedTx.Origin,
			}

			receipt.Paid = new(big.Int).Mul(new(big.Int).SetUint64(receipt.GasUsed), gasPrice)

			returnGas(leftOverGas)

			// reward
			//rewardRatio := builtin.Params.Native(rt.state).Get(polo.KeyRewardRatio)
			//overallGasPrice := tx.OverallGasPrice(baseGasPrice, rt.ctx.Number-1, rt.Seeker().GetID)
			//overallGasPrice := baseGasPrice

			reward := new(big.Int).SetUint64(receipt.GasUsed)
			reward.Mul(reward, gasPrice)
			//reward.Mul(reward, rewardRatio)
			//reward.Div(reward, big.NewInt(1e18))
			rt.state.AddBalance(rt.ctx.Beneficiary, reward)
			//builtin.Energy.Native(rt.state, rt.ctx.Time).Add()

			receipt.Reward = reward
			return receipt, nil
		},
	}, nil
}
