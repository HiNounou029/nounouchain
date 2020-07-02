// Copyright (c) 2019 The PoloChain developers

// Distributed under the GNU Lesser General Public License v3.0 software license, see the accompanying
// file LICENSE or <https://www.gnu.org/licenses/lgpl-3.0.html>

//+build gm

package main

import "github.com/ethereum/go-ethereum/p2p/discover"

// 国密public key
var bootstrapNodes = []*discover.Node{
	// first node as seed
	discover.MustParseNode("enode://dff0c5539a9d05f680e66aff57f40440b51aaf9d7ef89bda26ef7ce33a99985c164e07a181428c8e32fa6e13e59a8aa9011b59dd2d0677edde2ae493924a2db0@127.0.0.1:11235"),

	// separate node as seed
	discover.MustParseNode("enode://664ee1427dc60b15e9224257e29d88598e562e5dfe34e257fc2f59451fa90ef690f62ab772797501500c32250d6a178e4189cad604118d233eeec2b4b2d65469@127.0.0.1:55555"),
}
