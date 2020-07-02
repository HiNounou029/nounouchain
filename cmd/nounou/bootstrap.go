// Copyright (c) 2019 The PoloChain developers

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>

//+build !gm

package main

import "github.com/ethereum/go-ethereum/p2p/discover"

//非国密
var bootstrapNodes = []*discover.Node{
	// first node as seed
	discover.MustParseNode("enode://59b7dd6c3a43bb3bfbee7bed48da11b5d73a5955f70bd2f5fea370c7e89861a228137b9e48bde4065eff88a8121a842f091ab77cefd98c02bc53224494f9fbdd@127.0.0.1:11235"),

	// separate node as seed, see seed key from /data/seed.key
	discover.MustParseNode("enode://adbe99b8382dff3235b5571007a39b949755afeb4e1b777c1c4684177a419616f2bee53139721681cf258d640c8ff21eed1d0b0e7344c9ef160ef99ff6f8e74a@127.0.0.1:55555"),
}
