// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package eventslegacy

import (
	"context"
	"net/http"

	"github.com/HiNounou029/nounouchain/api/utils"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type EventsLegacy struct {
	db *logdb.LogDB
}

func New(db *logdb.LogDB) *EventsLegacy {
	return &EventsLegacy{
		db,
	}
}

//Filter query events with option
func (e *EventsLegacy) filter(ctx context.Context, filter *FilterLegacy) ([]*FilteredEvent, error) {
	f := convertFilter(filter)
	events, err := e.db.FilterEvents(ctx, f)
	if err != nil {
		return nil, err
	}
	fes := make([]*FilteredEvent, len(events))
	for i, e := range events {
		fes[i] = convertEvent(e)
	}
	return fes, nil
}

func (e *EventsLegacy) handleFilter(w http.ResponseWriter, req *http.Request) error {
	var filter FilterLegacy
	if err := utils.ParseJSON(req.Body, &filter); err != nil {
		return utils.BadRequest(errors.WithMessage(err, "body"))
	}
	query := req.URL.Query()
	if query.Get("address") != "" {
		addr, err := polo.ParseAddress(query.Get("address"))
		if err != nil {
			return utils.BadRequest(errors.WithMessage(err, "address"))
		}
		filter.Address = &addr
	}
	order := query.Get("order")
	if order != string(logdb.DESC) {
		filter.Order = logdb.ASC
	} else {
		filter.Order = logdb.DESC
	}
	fes, err := e.filter(req.Context(), &filter)
	if err != nil {
		return err
	}
	return utils.WriteJSON(w, fes)
}

func (e *EventsLegacy) Mount(root *mux.Router, pathPrefix string) {
	sub := root.PathPrefix(pathPrefix).Subrouter()

	sub.Path("").Methods("POST").HandlerFunc(utils.WrapHandlerFunc(e.handleFilter))
}
