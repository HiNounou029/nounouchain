// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package builtin

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/ethereum/go-ethereum/common"
)

func init() {
	defines := []struct {
		name string
		run  func(env *xenv.Environment) []interface{}
	}{
		{"native_executor", func(env *xenv.Environment) []interface{} {
			env.UseGas(polo.SloadGas)
			addr := polo.BytesToAddress(Params.Native(env.State()).Get(polo.KeyExecutorAddress).Bytes())
			return []interface{}{addr}
		}},
		{"native_add", func(env *xenv.Environment) []interface{} {
			var args struct {
				NodeMaster common.Address
				Endorsor   common.Address
				Identity   common.Hash
			}
			env.ParseArgs(&args)

			env.UseGas(polo.SloadGas)
			ok := Authority.Native(env.State()).Add(
				polo.Address(args.NodeMaster),
				polo.Address(args.Endorsor),
				polo.Bytes32(args.Identity))

			if ok {
				env.UseGas(polo.SstoreSetGas)
				env.UseGas(polo.SstoreResetGas)
			}
			return []interface{}{ok}
		}},
		{"native_revoke", func(env *xenv.Environment) []interface{} {
			var nodeMaster common.Address
			env.ParseArgs(&nodeMaster)

			env.UseGas(polo.SloadGas)
			ok := Authority.Native(env.State()).Revoke(polo.Address(nodeMaster))
			if ok {
				env.UseGas(polo.SstoreResetGas * 3)
			}
			return []interface{}{ok}
		}},
		{"native_get", func(env *xenv.Environment) []interface{} {
			var nodeMaster common.Address
			env.ParseArgs(&nodeMaster)

			env.UseGas(polo.SloadGas * 2)
			listed, endorsor, identity, active := Authority.Native(env.State()).Get(polo.Address(nodeMaster))

			return []interface{}{listed, endorsor, identity, active}
		}},
		{"native_first", func(env *xenv.Environment) []interface{} {
			env.UseGas(polo.SloadGas)
			if nodeMaster := Authority.Native(env.State()).First(); nodeMaster != nil {
				return []interface{}{*nodeMaster}
			}
			return []interface{}{polo.Address{}}
		}},
		{"native_next", func(env *xenv.Environment) []interface{} {
			var nodeMaster common.Address
			env.ParseArgs(&nodeMaster)

			env.UseGas(polo.SloadGas)
			if next := Authority.Native(env.State()).Next(polo.Address(nodeMaster)); next != nil {
				return []interface{}{*next}
			}
			return []interface{}{polo.Address{}}
		}},
		{"native_isEndorsed", func(env *xenv.Environment) []interface{} {
			var nodeMaster common.Address
			env.ParseArgs(&nodeMaster)

			env.UseGas(polo.SloadGas * 2)
			listed, endorsor, _, _ := Authority.Native(env.State()).Get(polo.Address(nodeMaster))
			if !listed {
				return []interface{}{false}
			}

			env.UseGas(polo.GetBalanceGas)
			bal := env.State().GetBalance(endorsor)

			env.UseGas(polo.SloadGas)
			endorsement := Params.Native(env.State()).Get(polo.KeyProposerEndorsement)
			return []interface{}{bal.Cmp(endorsement) >= 0}
		}},
	}
	abi := Authority.NativeABI()
	for _, def := range defines {
		if method, found := abi.MethodByName(def.name); found {
			nativeMethods[methodKey{Authority.Address, method.ID()}] = &nativeMethod{
				abi: method,
				run: def.run,
			}
		} else {
			panic("method not found: " + def.name)
		}
	}
}
