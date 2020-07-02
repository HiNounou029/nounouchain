// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package builtin

import (
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/abi"
	"github.com/HiNounou029/nounouchain/nounou/builtin/authority"
	"github.com/HiNounou029/nounouchain/nounou/builtin/gen"
	"github.com/HiNounou029/nounouchain/nounou/builtin/params"
	"github.com/HiNounou029/nounouchain/nounou/builtin/prototype"
	"github.com/HiNounou029/nounouchain/common/xenv"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/pkg/errors"
)

// Builtin contracts binding.
var (
	Params    = &paramsContract{mustLoadContract("Params")}
	Authority = &authorityContract{mustLoadContract("Authority")}
	Executor  = &executorContract{mustLoadContract("Executor")}
	Prototype = &prototypeContract{mustLoadContract("Prototype")}
	Extension = &extensionContract{mustLoadContract("Extension")}
	Token     = &tokenContract{mustLoadContract("Token")}
	Measure   = mustLoadContract("Measure")
)

type (
	paramsContract    struct{ *contract }
	authorityContract struct{ *contract }
	executorContract  struct{ *contract }
	prototypeContract struct{ *contract }
	extensionContract struct{ *contract }
	tokenContract     struct{ *contract }
)

func (p *paramsContract) Native(state *state.State) *params.Params {
	return params.New(p.Address, state)
}

func (a *authorityContract) Native(state *state.State) *authority.Authority {
	return authority.New(a.Address, state)
}

func (p *prototypeContract) Native(state *state.State) *prototype.Prototype {
	return prototype.New(p.Address, state)
}

func (p *prototypeContract) Events() *abi.ABI {
	asset := "compiled/PrototypeEvent.abi"
	data := gen.MustAsset(asset)
	abi, err := abi.New(data)
	if err != nil {
		panic(errors.Wrap(err, "load ABI for "+asset))
	}
	return abi
}

type nativeMethod struct {
	abi *abi.Method
	run func(env *xenv.Environment) []interface{}
}

type methodKey struct {
	polo.Address
	abi.MethodID
}

var nativeMethods = make(map[methodKey]*nativeMethod)

// FindNativeCall find native calls.
func FindNativeCall(to polo.Address, input []byte) (*abi.Method, func(*xenv.Environment) []interface{}, bool) {
	methodID, err := abi.ExtractMethodID(input)
	if err != nil {
		return nil, nil, false
	}

	method := nativeMethods[methodKey{to, methodID}]
	if method == nil {
		return nil, nil, false
	}
	return method.abi, method.run, true
}
