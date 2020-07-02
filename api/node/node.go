// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package node

import (
	"net/http"

	"github.com/HiNounou029/nounouchain/api/utils"
	"github.com/gorilla/mux"
)

type Node struct {
	nw Network
}

func New(nw Network) *Node {
	return &Node{
		nw,
	}
}

func (n *Node) PeersStats() []*PeerStats {
	return ConvertPeersStats(n.nw.PeersStats())
}

func (n *Node) handleNetwork(w http.ResponseWriter, req *http.Request) error {
	return utils.WriteTo(w, req, n.PeersStats())
}

func (n *Node) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()

	sub.Path("/network/peers").Methods("Get").HandlerFunc(utils.WrapHandlerFunc(n.handleNetwork))
}
