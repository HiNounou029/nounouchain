// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package status

import (
	"github.com/HiNounou029/nounouchain/api/utils"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/gorilla/mux"
	"net/http"
)

type Status struct {
	chain *chain.Chain
}

func New(chain *chain.Chain) *Status {
	return &Status{
		chain,
	}
}

func (s *Status) handleGetStatus(w http.ResponseWriter, req *http.Request) error {
	cs := convertChainStatus(s.chain)

	return utils.WriteTo(w, req, cs)
}

func (s *Status) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()
	sub.Path("/").Methods("GET").HandlerFunc(utils.WrapHandlerFunc(s.handleGetStatus))

}
