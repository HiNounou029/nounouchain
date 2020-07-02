// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package api

import (
	"github.com/HiNounou029/nounouchain/api/authority"
	"github.com/HiNounou029/nounouchain/api/cert"
	"net/http"
	"strings"

	"github.com/HiNounou029/nounouchain/api/accounts"
	"github.com/HiNounou029/nounouchain/api/blocks"
	"github.com/HiNounou029/nounouchain/api/events"
	"github.com/HiNounou029/nounouchain/api/eventslegacy"
	"github.com/HiNounou029/nounouchain/api/node"
	"github.com/HiNounou029/nounouchain/api/status"
	"github.com/HiNounou029/nounouchain/api/subscriptions"
	"github.com/HiNounou029/nounouchain/api/transactions"
	"github.com/HiNounou029/nounouchain/api/transfers"
	"github.com/HiNounou029/nounouchain/api/transferslegacy"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/HiNounou029/nounouchain/state"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

var (
	ApiVer = "1.0.0"
)

//New return api router
func New(chain *chain.Chain, stateCreator *state.Creator, txPool *txpool.TxPool,
	logDB *logdb.LogDB, nw node.Network, allowedOrigins string,
	backtraceLimit uint32, callGasLimit uint64, path string) (http.HandlerFunc, func()) {
	origins := strings.Split(strings.TrimSpace(allowedOrigins), ",")
	for i, o := range origins {
		origins[i] = strings.ToLower(strings.TrimSpace(o))
	}

	router := mux.NewRouter()

	accounts.New(chain, stateCreator, callGasLimit).
		Mount(router, "/accounts")
	eventslegacy.New(logDB).
		Mount(router, "/events")
	transferslegacy.New(logDB).
		Mount(router, "/transfers")
	eventslegacy.New(logDB).
		Mount(router, "/logs/events")
	events.New(logDB).
		Mount(router, "/logs/event")
	transferslegacy.New(logDB).
		Mount(router, "/logs/transfers")
	transfers.New(logDB).
		Mount(router, "/logs/transfer")
	blocks.New(chain).
		Mount(router, "/blocks")
	status.New(chain).
		Mount(router, "/status")
	transactions.New(chain, txPool).
		Mount(router, "/transactions")
	node.New(nw).
		Mount(router, "/node")
	authority.New(chain, stateCreator).
		Mount(router, "/authority")

	cert.New(path).Mount(router, "/verify")

	subs := subscriptions.New(chain, origins, backtraceLimit)
	subs.Mount(router, "/subscriptions")

	return handlers.CORS(
			handlers.AllowedOrigins(origins),
			handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
			handlers.AllowedHeaders([]string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "x-genesis-id"}),
            handlers.AllowCredentials())(router).ServeHTTP,
		subs.Close // subscriptions handles hijacked conns, which need to be closed
}
