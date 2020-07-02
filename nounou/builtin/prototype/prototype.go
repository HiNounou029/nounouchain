// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package prototype

import (
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/ethereum/go-ethereum/rlp"
)

type Prototype struct {
	addr  polo.Address
	state *state.State
}

func New(addr polo.Address, state *state.State) *Prototype {
	return &Prototype{addr, state}
}

func (p *Prototype) Bind(self polo.Address) *Binding {
	return &Binding{p.addr, p.state, self}
}

type Binding struct {
	addr  polo.Address
	state *state.State
	self  polo.Address
}

func (b *Binding) userKey(user polo.Address) polo.Bytes32 {
	return polo.Blake2b(b.self.Bytes(), user.Bytes(), []byte("user"))
}

func (b *Binding) creditPlanKey() polo.Bytes32 {
	return polo.Blake2b(b.self.Bytes(), []byte("credit-plan"))
}

func (b *Binding) sponsorKey(sponsor polo.Address) polo.Bytes32 {
	return polo.Blake2b(b.self.Bytes(), sponsor.Bytes(), []byte("sponsor"))
}

func (b *Binding) curSponsorKey() polo.Bytes32 {
	return polo.Blake2b(b.self.Bytes(), []byte("cur-sponsor"))
}

func (b *Binding) getUserObject(user polo.Address) *userObject {
	var uo userObject
	b.state.DecodeStorage(b.addr, b.userKey(user), func(raw []byte) error {
		if len(raw) == 0 {
			uo = userObject{&big.Int{}, 0}
			return nil
		}
		return rlp.DecodeBytes(raw, &uo)
	})
	return &uo
}

func (b *Binding) setUserObject(user polo.Address, uo *userObject) {
	b.state.EncodeStorage(b.addr, b.userKey(user), func() ([]byte, error) {
		if uo.IsEmpty() {
			return nil, nil
		}
		return rlp.EncodeToBytes(uo)
	})
}

func (b *Binding) getCreditPlan() *creditPlan {
	var cp creditPlan
	b.state.DecodeStorage(b.addr, b.creditPlanKey(), func(raw []byte) error {
		if len(raw) == 0 {
			cp = creditPlan{&big.Int{}, &big.Int{}}
			return nil
		}
		return rlp.DecodeBytes(raw, &cp)
	})
	return &cp
}

func (b *Binding) setCreditPlan(cp *creditPlan) {
	b.state.EncodeStorage(b.addr, b.creditPlanKey(), func() ([]byte, error) {
		if cp.IsEmpty() {
			return nil, nil
		}
		return rlp.EncodeToBytes(cp)
	})
}

func (b *Binding) IsUser(user polo.Address) bool {
	return !b.getUserObject(user).IsEmpty()
}

func (b *Binding) AddUser(user polo.Address, blockTime uint64) {
	b.setUserObject(user, &userObject{&big.Int{}, blockTime})
}

func (b *Binding) RemoveUser(user polo.Address) {
	// set to empty
	b.setUserObject(user, &userObject{&big.Int{}, 0})
}

func (b *Binding) UserCredit(user polo.Address, blockTime uint64) *big.Int {
	uo := b.getUserObject(user)
	if uo.IsEmpty() {
		return &big.Int{}
	}
	return uo.Credit(b.getCreditPlan(), blockTime)
}

func (b *Binding) SetUserCredit(user polo.Address, credit *big.Int, blockTime uint64) {
	up := b.getCreditPlan()
	used := new(big.Int).Sub(up.Credit, credit)
	if used.Sign() < 0 {
		used = &big.Int{}
	}
	b.setUserObject(user, &userObject{used, blockTime})
}

func (b *Binding) CreditPlan() (credit, recoveryRate *big.Int) {
	cp := b.getCreditPlan()
	return cp.Credit, cp.RecoveryRate
}

func (b *Binding) SetCreditPlan(credit, recoveryRate *big.Int) {
	b.setCreditPlan(&creditPlan{credit, recoveryRate})
}

func (b *Binding) Sponsor(sponsor polo.Address, flag bool) {
	b.state.EncodeStorage(b.addr, b.sponsorKey(sponsor), func() ([]byte, error) {
		if !flag {
			return nil, nil
		}
		return rlp.EncodeToBytes(&flag)
	})
}

func (b *Binding) IsSponsor(sponsor polo.Address) (flag bool) {
	b.state.DecodeStorage(b.addr, b.sponsorKey(sponsor), func(raw []byte) error {
		if len(raw) == 0 {
			return nil
		}
		return rlp.DecodeBytes(raw, &flag)
	})
	return
}

func (b *Binding) SelectSponsor(sponsor polo.Address) {
	b.state.SetStorage(b.addr, b.curSponsorKey(), polo.BytesToBytes32(sponsor.Bytes()))
}

func (b *Binding) CurrentSponsor() polo.Address {
	return polo.BytesToAddress(b.state.GetStorage(b.addr, b.curSponsorKey()).Bytes())
}
