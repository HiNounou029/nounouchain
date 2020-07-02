// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package transferslegacy

import (
	"context"
	"net/http"

	"github.com/HiNounou029/nounouchain/api/utils"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type TransfersLegacy struct {
	db *logdb.LogDB
}

func New(db *logdb.LogDB) *TransfersLegacy {
	return &TransfersLegacy{
		db,
	}
}

//Filter query logs with option
func (t *TransfersLegacy) filter(ctx context.Context, filter *logdb.TransferFilter) ([]*FilteredTransfer, error) {
	transfers, err := t.db.FilterTransfers(ctx, filter)
	if err != nil {
		return nil, err
	}
	tLogs := make([]*FilteredTransfer, len(transfers))
	for i, trans := range transfers {
		tLogs[i] = convertTransfer(trans)
	}
	return tLogs, nil
}

func (t *TransfersLegacy) handleFilterTransferLogs(w http.ResponseWriter, req *http.Request) error {
	var filter TransferFilter
	if err := utils.ParseJSON(req.Body, &filter); err != nil {
		return utils.BadRequest(errors.WithMessage(err, "body"))
	}
	order := req.URL.Query().Get("order")
	if order != string(logdb.DESC) {
		filter.Order = logdb.ASC
	} else {
		filter.Order = logdb.DESC
	}
	tLogs, err := t.filter(req.Context(), convertTransferFilter(&filter))
	if err != nil {
		return err
	}
	return utils.WriteJSON(w, tLogs)
}

func (t *TransfersLegacy) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()

	sub.Path("").Methods("POST").HandlerFunc(utils.WrapHandlerFunc(t.handleFilterTransferLogs))
}
