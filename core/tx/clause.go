// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package tx

import (
	"fmt"
	"io"
	"math/big"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/ethereum/go-ethereum/rlp"
)

type clauseBody struct {
	To    *polo.Address `rlp:"nil"`
	Value *big.Int
	Data  []byte
}

// Clause is the basic execution unit of a transaction.
type Clause struct {
	body clauseBody
}

// NewClause create a new clause instance.
func NewClause(to *polo.Address) *Clause {
	if to != nil {
		// make a copy of 'to'
		cpy := *to
		to = &cpy
	}
	return &Clause{
		clauseBody{
			to,
			&big.Int{},
			nil,
		},
	}
}

// WithValue create a new clause copy with value changed.
func (c *Clause) WithValue(value *big.Int) *Clause {
	newClause := *c
	newClause.body.Value = new(big.Int).Set(value)
	return &newClause
}

// WithData create a new clause copy with data changed.
func (c *Clause) WithData(data []byte) *Clause {
	newClause := *c
	newClause.body.Data = append([]byte(nil), data...)
	return &newClause
}

// To returns 'To' address.
func (c *Clause) To() *polo.Address {
	if c.body.To == nil {
		return nil
	}
	cpy := *c.body.To
	return &cpy
}

// Value returns 'Value'.
func (c *Clause) Value() *big.Int {
	return new(big.Int).Set(c.body.Value)
}

// Data returns 'Data'.
func (c *Clause) Data() []byte {
	return append([]byte(nil), c.body.Data...)
}

// IsCreatingContract return if this clause is going to create a contract.
func (c *Clause) IsCreatingContract() bool {
	return c.body.To == nil
}

// EncodeRLP implements rlp.Encoder
func (c *Clause) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &c.body)
}

// DecodeRLP implements rlp.Decoder
func (c *Clause) DecodeRLP(s *rlp.Stream) error {
	var body clauseBody
	if err := s.Decode(&body); err != nil {
		return err
	}
	*c = Clause{body}
	return nil
}

func (c *Clause) String() string {
	var to string
	if c.body.To == nil {
		to = "nil"
	} else {
		to = c.body.To.String()
	}
	return fmt.Sprintf(`
		(To:	%v
		 Value:	%v
		 Data:	0x%x)`, to, c.body.Value, c.body.Data)
}
