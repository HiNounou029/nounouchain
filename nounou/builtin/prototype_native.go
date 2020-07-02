// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package builtin

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/ethereum/go-ethereum/common"
)

func init() {

	events := Prototype.Events()

	mustEventByName := func(name string) *abi.Event {
		if event, found := events.EventByName(name); found {
			return event
		}
		panic("event not found")
	}

	masterEvent := mustEventByName("$Master")
	creditPlanEvent := mustEventByName("$CreditPlan")
	userEvent := mustEventByName("$User")
	sponsorEvent := mustEventByName("$Sponsor")

	defines := []struct {
		name string
		run  func(env *xenv.Environment) []interface{}
	}{
		{"native_master", func(env *xenv.Environment) []interface{} {
			var self common.Address
			env.ParseArgs(&self)

			env.UseGas(polo.GetBalanceGas)
			master := env.State().GetMaster(polo.Address(self))

			return []interface{}{master}
		}},
		{"native_setMaster", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self      common.Address
				NewMaster common.Address
			}
			env.ParseArgs(&args)

			env.UseGas(polo.SstoreResetGas)
			env.State().SetMaster(polo.Address(args.Self), polo.Address(args.NewMaster))

			env.Log(masterEvent, polo.Address(args.Self), nil, args.NewMaster)
			return nil
		}},
		{"native_balanceAtBlock", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self        common.Address
				BlockNumber uint32
			}
			env.ParseArgs(&args)
			ctx := env.BlockContext()

			if args.BlockNumber > ctx.Number {
				return []interface{}{&big.Int{}}
			}

			if ctx.Number-args.BlockNumber > polo.MaxBackTrackingBlockNumber {
				return []interface{}{&big.Int{}}
			}

			if args.BlockNumber == ctx.Number {
				env.UseGas(polo.GetBalanceGas)
				val := env.State().GetBalance(polo.Address(args.Self))
				return []interface{}{val}
			}

			env.UseGas(polo.SloadGas)
			blockID := env.Seeker().GetID(args.BlockNumber)

			env.UseGas(polo.SloadGas)
			header := env.Seeker().GetHeader(blockID)

			env.UseGas(polo.SloadGas)
			state := env.State().Spawn(header.StateRoot())

			env.UseGas(polo.GetBalanceGas)
			val := state.GetBalance(polo.Address(args.Self))

			return []interface{}{val}
		}},
		{"native_hasCode", func(env *xenv.Environment) []interface{} {
			var self common.Address
			env.ParseArgs(&self)

			env.UseGas(polo.GetBalanceGas)
			hasCode := !env.State().GetCodeHash(polo.Address(self)).IsZero()

			return []interface{}{hasCode}
		}},
		{"native_storageFor", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self common.Address
				Key  polo.Bytes32
			}
			env.ParseArgs(&args)

			env.UseGas(polo.SloadGas)
			storage := env.State().GetStorage(polo.Address(args.Self), args.Key)
			return []interface{}{storage}
		}},
		{"native_creditPlan", func(env *xenv.Environment) []interface{} {
			var self common.Address
			env.ParseArgs(&self)
			binding := Prototype.Native(env.State()).Bind(polo.Address(self))

			env.UseGas(polo.SloadGas)
			credit, rate := binding.CreditPlan()

			return []interface{}{credit, rate}
		}},
		{"native_setCreditPlan", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self         common.Address
				Credit       *big.Int
				RecoveryRate *big.Int
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SstoreSetGas)
			binding.SetCreditPlan(args.Credit, args.RecoveryRate)
			env.Log(creditPlanEvent, polo.Address(args.Self), nil, args.Credit, args.RecoveryRate)
			return nil
		}},
		{"native_isUser", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self common.Address
				User common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SloadGas)
			isUser := binding.IsUser(polo.Address(args.User))

			return []interface{}{isUser}
		}},
		{"native_userCredit", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self common.Address
				User common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(2 * polo.SloadGas)
			credit := binding.UserCredit(polo.Address(args.User), env.BlockContext().Time)

			return []interface{}{credit}
		}},
		{"native_addUser", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self common.Address
				User common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SloadGas)
			if binding.IsUser(polo.Address(args.User)) {
				return []interface{}{false}
			}

			env.UseGas(polo.SstoreSetGas)
			binding.AddUser(polo.Address(args.User), env.BlockContext().Time)

			var action polo.Bytes32
			copy(action[:], "added")
			env.Log(userEvent, polo.Address(args.Self), []polo.Bytes32{polo.BytesToBytes32(args.User[:])}, action)
			return []interface{}{true}
		}},
		{"native_removeUser", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self common.Address
				User common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SloadGas)
			if !binding.IsUser(polo.Address(args.User)) {
				return []interface{}{false}
			}

			env.UseGas(polo.SstoreResetGas)
			binding.RemoveUser(polo.Address(args.User))

			var action polo.Bytes32
			copy(action[:], "removed")
			env.Log(userEvent, polo.Address(args.Self), []polo.Bytes32{polo.BytesToBytes32(args.User[:])}, action)
			return []interface{}{true}
		}},
		{"native_sponsor", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self    common.Address
				Sponsor common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SloadGas)
			if binding.IsSponsor(polo.Address(args.Sponsor)) {
				return []interface{}{false}
			}

			env.UseGas(polo.SstoreSetGas)
			binding.Sponsor(polo.Address(args.Sponsor), true)

			var action polo.Bytes32
			copy(action[:], "sponsored")
			env.Log(sponsorEvent, polo.Address(args.Self), []polo.Bytes32{polo.BytesToBytes32(args.Sponsor.Bytes())}, action)
			return []interface{}{true}
		}},
		{"native_unsponsor", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self    common.Address
				Sponsor common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SloadGas)
			if !binding.IsSponsor(polo.Address(args.Sponsor)) {
				return []interface{}{false}
			}

			env.UseGas(polo.SstoreResetGas)
			binding.Sponsor(polo.Address(args.Sponsor), false)

			var action polo.Bytes32
			copy(action[:], "unsponsored")
			env.Log(sponsorEvent, polo.Address(args.Self), []polo.Bytes32{polo.BytesToBytes32(args.Sponsor.Bytes())}, action)
			return []interface{}{true}
		}},
		{"native_isSponsor", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self    common.Address
				Sponsor common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SloadGas)
			isSponsor := binding.IsSponsor(polo.Address(args.Sponsor))

			return []interface{}{isSponsor}
		}},
		{"native_selectSponsor", func(env *xenv.Environment) []interface{} {
			var args struct {
				Self    common.Address
				Sponsor common.Address
			}
			env.ParseArgs(&args)
			binding := Prototype.Native(env.State()).Bind(polo.Address(args.Self))

			env.UseGas(polo.SloadGas)
			if !binding.IsSponsor(polo.Address(args.Sponsor)) {
				return []interface{}{false}
			}

			env.UseGas(polo.SstoreResetGas)
			binding.SelectSponsor(polo.Address(args.Sponsor))

			var action polo.Bytes32
			copy(action[:], "selected")
			env.Log(sponsorEvent, polo.Address(args.Self), []polo.Bytes32{polo.BytesToBytes32(args.Sponsor.Bytes())}, action)

			return []interface{}{true}
		}},
		{"native_currentSponsor", func(env *xenv.Environment) []interface{} {
			var self common.Address
			env.ParseArgs(&self)
			binding := Prototype.Native(env.State()).Bind(polo.Address(self))

			env.UseGas(polo.SloadGas)
			addr := binding.CurrentSponsor()

			return []interface{}{addr}
		}},
	}
	abi := Prototype.NativeABI()
	for _, def := range defines {
		if method, found := abi.MethodByName(def.name); found {
			nativeMethods[methodKey{Prototype.Address, method.ID()}] = &nativeMethod{
				abi: method,
				run: def.run,
			}
		} else {
			panic("method not found: " + def.name)
		}
	}
}
