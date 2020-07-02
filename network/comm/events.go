// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package comm

import (
	"context"

	"github.com/HiNounou029/nounouchain/core/block"
)

// NewBlockEvent event emitted when received block announcement.
type NewBlockEvent struct {
	*block.Block
}

// HandleBlockStream to handle the stream of downloaded blocks in sync process.
type HandleBlockStream func(ctx context.Context, stream <-chan *block.Block) error
