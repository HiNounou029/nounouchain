// Copyright © 2018-2019 Apollo Technologies Pte. Ltd. All Rights Reserved.

package comm

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/HiNounou029/nounouchain/polo"
	"github.com/HiNounou029/nounouchain/common/co"
	"github.com/HiNounou029/nounouchain/core/block"
	"github.com/HiNounou029/nounouchain/core/chain"
	"github.com/HiNounou029/nounouchain/core/tx"
	"github.com/HiNounou029/nounouchain/core/txpool"
	"github.com/HiNounou029/nounouchain/network"
	"github.com/HiNounou029/nounouchain/network/comm/proto"
	"github.com/ethereum/go-ethereum/event"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/inconshreveable/log15"

	. "github.com/HiNounou029/nounouchain/caclient"
	"github.com/HiNounou029/nounouchain/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
)

var log = log15.New("pkg", "comm")

// Communicator communicates with remote p2p peers to exchange blocks and txs, etc.
type Communicator struct {
	chain          *chain.Chain
	txPool         *txpool.TxPool
	ctx            context.Context
	cancel         context.CancelFunc
	peerSet        *PeerSet
	syncedCh       chan struct{}
	newBlockFeed   event.Feed
	announcementCh chan *announcement
	feedScope      event.SubscriptionScope
	goes           co.Goes
	onceSynced     sync.Once
	rootCaPath	   string
	isWithCert	   bool
	certInfo	   []byte
	nodeInfo *p2p.NodeInfo
}

// New create a new Communicator instance.
func New(chain *chain.Chain, txPool *txpool.TxPool, rootCaPath string, isStartWithCert bool, certInfo []byte) *Communicator {
	ctx, cancel := context.WithCancel(context.Background())
	return &Communicator{
		chain:          chain,
		txPool:         txPool,
		ctx:            ctx,
		cancel:         cancel,
		peerSet:        newPeerSet(),
		syncedCh:       make(chan struct{}),
		announcementCh: make(chan *announcement),
		rootCaPath:     rootCaPath,
		isWithCert:     isStartWithCert,
		certInfo: 		certInfo,
	}
}

// Synced returns a channel indicates if synchronization process passed.
func (c *Communicator) Synced() <-chan struct{} {
	return c.syncedCh
}

// Sync start synchronization process.
func (c *Communicator) Sync(handler HandleBlockStream) {
	const initSyncInterval = 2 * time.Second
//	const syncInterval = 30 * time.Second
	syncInterval := time.Duration(2*polo.Conf.BlockInterval) * time.Second

	c.goes.Go(func() {
		timer := time.NewTimer(0)
		defer timer.Stop()
		delay := initSyncInterval
		syncCount := 0

		shouldSynced := func() bool {
			bestBlockTime := c.chain.BestBlock().Header().Timestamp()
			now := uint64(time.Now().Unix())
			if bestBlockTime+polo.Conf.BlockInterval >= now {
				return true
			}
			if syncCount > 2 {
				return true
			}
			return false
		}

		for {
			timer.Stop()
			timer = time.NewTimer(delay)
			select {
			case <-c.ctx.Done():
				return
			case <-timer.C:
				log.Debug("synchronization start")

				best := c.chain.BestBlock().Header()
				// choose peer which has the head block with higher total score
				peer := c.peerSet.Slice().Find(func(peer *Peer) bool {
					_, totalScore := peer.Head()
					return totalScore >= best.TotalScore()
				})
				if peer == nil {
					if c.peerSet.Len() < 1 {
						log.Debug("no suitable peer to sync")
						break
					}
					// if more than 3 peers connected, we are assumed to be the best
					log.Debug("synchronization done, best assumed")
				} else {
					if err := c.sync(peer, best.Number(), handler); err != nil {
						peer.logger.Debug("synchronization failed", "err", err)
						break
					}
					peer.logger.Debug("synchronization done")
				}
				syncCount++

				if shouldSynced() {
					delay = syncInterval
					c.onceSynced.Do(func() {
						close(c.syncedCh)
					})
				}
			}
		}
	})
}

// Protocols returns all supported protocols.
func (c *Communicator) Protocols() []*network.Protocol {
	genesisID := c.chain.GenesisBlock().Header().ID()
	var buf bytes.Buffer
	fmt.Fprintf(&buf,"%v%v@%x@%x", proto.Name, proto.Version, genesisID[24:], polo.ConfigId[:8])
	fmt.Println(buf.String())
	return []*network.Protocol{
		&network.Protocol{
			Protocol: p2p.Protocol{
				Name:    proto.Name,
				Version: proto.Version,
				Length:  proto.Length,
				Run:     c.servePeer,
			},
//			DiscTopic: fmt.Sprintf("%v%v@%x", proto.Name, proto.Version, genesisID[24:]),
			DiscTopic: fmt.Sprintf("%v%v@%x@%x", proto.Name, proto.Version, genesisID[24:], polo.ConfigId[:8]),
		}}
}

// Start start the communicator.
func (c *Communicator) Start(nodeInfo *p2p.NodeInfo) {
	c.goes.Go(c.txsLoop)
	c.goes.Go(c.announcementLoop)
	c.nodeInfo = nodeInfo
}

// Stop stop the communicator.
func (c *Communicator) Stop() {
	c.cancel()
	c.feedScope.Close()
	c.goes.Wait()
}

type txsToSync struct {
	txs    tx.Transactions
	synced bool
}

func (c *Communicator) servePeer(p *p2p.Peer, rw p2p.MsgReadWriter) error {
	peer := newPeer(p, rw)
	c.goes.Go(func() {
		c.runPeer(peer)
	})

	var txsToSync txsToSync

	return peer.Serve(func(msg *p2p.Msg, w func(interface{})) error {
		return c.handleRPC(peer, msg, w, &txsToSync)
	}, proto.MaxMsgSize)
}

func (c *Communicator) runPeer(peer *Peer) {
	defer peer.Disconnect(p2p.DiscRequested)

	// 5sec timeout for handshake
	ctx, cancel := context.WithTimeout(c.ctx, time.Second*5)
	defer cancel()

	status, err := proto.GetStatus(ctx, peer)
	if err != nil {
		peer.logger.Debug("failed to get status", "err", err)
		return
	}
	if c.isWithCert == true{
		if len(status.CertInfo) == 0 {
			log.Info("there is no cert for this peer ！")
			proto.NotifyCertValRes(ctx, peer, "Please start with cert !")
			//c.peerSet.Remove(peer.ID())
			return
		}
		err = ValCert(c.rootCaPath, status.CertInfo)
		if err != nil {
			proto.NotifyCertValRes(ctx, peer, "Certificate validation failed ！")
			log.Info("Certificate validation failed ！")
			//c.peerSet.Remove(peer.ID())
			return
		}
	}
	if status.GenesisBlockID != c.chain.GenesisBlock().Header().ID() {
		peer.logger.Debug("failed to handshake", "err", "genesis id mismatch")
		return
	}
	localClock := uint64(time.Now().Unix())
	remoteClock := status.SysTimestamp

	diff := localClock - remoteClock
	if localClock < remoteClock {
		diff = remoteClock - localClock
	}
	if diff > polo.Conf.BlockInterval*2 {
		peer.logger.Debug("failed to handshake", "err", "sys time diff too large")
		return
	}

	peer.UpdateHead(status.BestBlockID, status.TotalScore)
	c.peerSet.Add(peer)
	peer.logger.Debug(fmt.Sprintf("peer added (%v)", c.peerSet.Len()))

	defer func() {
		c.peerSet.Remove(peer.ID())
		peer.logger.Debug(fmt.Sprintf("peer removed (%v)", c.peerSet.Len()))
	}()

	select {
	case <-peer.Done():
	case <-c.ctx.Done():
	case <-c.syncedCh:
		c.syncTxs(peer)
		select {
		case <-peer.Done():
		case <-c.ctx.Done():
		}
	}
}

// SubscribeBlock subscribe the event that new block received.
func (c *Communicator) SubscribeBlock(ch chan *NewBlockEvent) event.Subscription {
	return c.feedScope.Track(c.newBlockFeed.Subscribe(ch))
}

// BroadcastBlock broadcast a block to remote peers.
func (c *Communicator) BroadcastBlock(blk *block.Block) {
	peers := c.peerSet.Slice().Filter(func(p *Peer) bool {
		return !p.IsBlockKnown(blk.Header().ID())
	})

	p := int(math.Sqrt(float64(len(peers))))
	toPropagate := peers[:p]
	toAnnounce := peers[p:]

	for _, peer := range toPropagate {
		peer := peer
		peer.MarkBlock(blk.Header().ID())
		c.goes.Go(func() {
			if err := proto.NotifyNewBlock(c.ctx, peer, blk); err != nil {
				peer.logger.Debug("failed to broadcast new block", "err", err)
			}
		})
	}

	for _, peer := range toAnnounce {
		peer := peer
		peer.MarkBlock(blk.Header().ID())
		c.goes.Go(func() {

			if err := proto.NotifyNewBlockID(c.ctx, peer, blk.Header().ID()); err != nil {
				peer.logger.Debug("failed to broadcast new block id", "err", err)
			}
		})
	}
}

// PeerCount returns count of peers.
func (c *Communicator) PeerCount() int {
	return c.peerSet.Len()
}

// PeersStats returns all peers' stats
func (c *Communicator) PeersStats() []*PeerStats {
	var stats []*PeerStats
	for _, peer := range c.peerSet.Slice() {
		bestID, totalScore := peer.Head()
		pubkey, _ := peer.ID().Pubkey()
		stats = append(stats, &PeerStats{
			Name:        peer.Name(),
			BestBlockID: bestID,
			TotalScore:  totalScore,
			PeerID:      peer.ID().String(),
			NetAddr:     peer.RemoteAddr().String(),
			Inbound:     peer.Inbound(),
			Duration:    uint64(time.Duration(peer.Duration()) / time.Second),
			PeerAddress: crypto.PubkeyToAddress(*pubkey).String(),
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Duration < stats[j].Duration
	})

	if c.nodeInfo != nil {
		nodeID, _:= discover.HexID(c.nodeInfo.ID)
		pubkey, _ := nodeID.Pubkey()
		stats = append(stats, &PeerStats{
			Name:        c.nodeInfo.Name,
			BestBlockID: c.chain.BestBlock().Header().ID(),
			TotalScore:  c.chain.BestBlock().Header().TotalScore(),
			PeerID:      c.nodeInfo.ID,
			NetAddr:     c.nodeInfo.ListenAddr,
			Inbound:     true,
			Duration:    0,
			PeerAddress: crypto.PubkeyToAddress(*pubkey).String(),
		})
	}

	return stats
}
