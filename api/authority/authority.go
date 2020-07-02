package authority

import (
	"github.com/HiNounou029/nounouchain/api/utils"
	"github.com/HiNounou029/nounouchain/nounou/builtin"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/gorilla/mux"
	"math/big"
	"net/http"
)

type authority struct {
	chain        *chain.Chain
	stateCreator *state.Creator
}


func New(chain *chain.Chain, stateCreator *state.Creator) *authority {
	return &authority{
		chain,
		stateCreator,
	}
}

func (n *authority)handleAuthority(w http.ResponseWriter, req *http.Request)  error{
	header := n.chain.BestBlock().Header()
	state, err := n.stateCreator.NewState(header.StateRoot())
	if err != nil {
		return err
	}

	a := builtin.Authority.Native(state)
	endorsement := big.NewInt(0)
	candidates := a.Candidates(endorsement, 100)
	return utils.WriteTo(w, req, candidates)
}

func (n *authority) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()

	sub.Path("").Methods("Get").HandlerFunc(utils.WrapHandlerFunc(n.handleAuthority))
}