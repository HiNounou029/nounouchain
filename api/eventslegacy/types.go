// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package eventslegacy

import (
	"fmt"

	"github.com/HiNounou029/nounouchain/api/transactions"
	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/nounou/logdb"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type TopicSet struct {
	Topic0 *polo.Bytes32 `json:"topic0"`
	Topic1 *polo.Bytes32 `json:"topic1"`
	Topic2 *polo.Bytes32 `json:"topic2"`
	Topic3 *polo.Bytes32 `json:"topic3"`
	Topic4 *polo.Bytes32 `json:"topic4"`
}

type FilterLegacy struct {
	Address   *polo.Address
	TopicSets []*TopicSet
	Range     *logdb.Range
	Options   *logdb.Options
	Order     logdb.Order
}

func convertFilter(filter *FilterLegacy) *logdb.EventFilter {
	f := &logdb.EventFilter{
		Range:   filter.Range,
		Options: filter.Options,
		Order:   filter.Order,
	}
	if len(filter.TopicSets) > 0 {
		criterias := make([]*logdb.EventCriteria, len(filter.TopicSets))
		for i, topicSet := range filter.TopicSets {
			var topics [5]*polo.Bytes32
			topics[0] = topicSet.Topic0
			topics[1] = topicSet.Topic1
			topics[2] = topicSet.Topic2
			topics[3] = topicSet.Topic3
			topics[4] = topicSet.Topic4
			criteria := &logdb.EventCriteria{
				Address: filter.Address,
				Topics:  topics,
			}
			criterias[i] = criteria
		}
		f.CriteriaSet = criterias
	} else if filter.Address != nil {
		f.CriteriaSet = []*logdb.EventCriteria{
			&logdb.EventCriteria{
				Address: filter.Address,
			}}
	}
	return f
}

// FilteredEvent only comes from one contract
type FilteredEvent struct {
	Address polo.Address         `json:"address"`
	Topics  []*polo.Bytes32      `json:"topics"`
	Data    string               `json:"data"`
	Meta    transactions.LogMeta `json:"meta"`
}

//convert a logdb.Event into a json format Event
func convertEvent(event *logdb.Event) *FilteredEvent {
	fe := FilteredEvent{
		Address: event.Address,
		Data:    hexutil.Encode(event.Data),
		Meta: transactions.LogMeta{
			BlockID:        event.BlockID,
			BlockNumber:    event.BlockNumber,
			BlockTimestamp: event.BlockTime,
			TxID:           event.TxID,
			TxOrigin:       event.TxOrigin,
		},
	}
	fe.Topics = make([]*polo.Bytes32, 0)
	for i := 0; i < 5; i++ {
		if event.Topics[i] != nil {
			fe.Topics = append(fe.Topics, event.Topics[i])
		}
	}
	return &fe
}

func (e *FilteredEvent) String() string {
	return fmt.Sprintf(`
		Event(
			address: 	   %v,
			topics:        %v,
			data:          %v,
			meta: (blockID     %v,
				blockNumber    %v,
				blockTimestamp %v),
				txID     %v,
				txOrigin %v)
			)`,
		e.Address,
		e.Topics,
		e.Data,
		e.Meta.BlockID,
		e.Meta.BlockNumber,
		e.Meta.BlockTimestamp,
		e.Meta.TxID,
		e.Meta.TxOrigin,
	)
}
