# priv-libp2p-node

**WIP**

Libp2p node abstraction, it encapsulates a recipe of libp2p's common protocols / concepts 
(pubsub, dht, ipld, etc...).

Inspired by [ipfs-lite](https://github.com/hsanjuan/ipfs-lite), which is an alternative to a full IPFS, 
and comes by default with IPLD (DAGService).

This package makes it easy to configure several types of libp2p nodes:
- LibP2PNode (`./core/node.go`) is the most minimal interface
    - local datastore, DHT, pubsub
    - a [circuit-relay](https://docs.libp2p.io/concepts/circuit-relay/) node is an example (see `./cmd/relayer`)
- IpldNode (`./ipld/ipld_node.go`) - LibP2PNode + IPLD, i.e. more similar to ipfs-lite
    - it is a [ipld.DAGService](https://godoc.org/github.com/ipfs/go-ipld-format#DAGService)
    - [crdt](https://github.com/ipfs/go-ds-crdt) 
    could be configured easily to provide state consistency

## Install

As a library:

```bash
go get github.com/amirylm/priv-libp2p-node
```

Libp2p version -> `github.com/libp2p/go-libp2p@v0.11.0`

## Usage

See `./cmd` folder and tests for more concrete examples.

```go
package main

import (
    "log"

	p2pnode "github.com/amirylm/priv-libp2p-node/core"
	
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

func startNode(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (p2pnode.LibP2PNode, error) {
    node := p2pnode.NewBaseNode(context.Background(),
		p2pnode.NewConfig(priv, psk, nil),
		p2pnode.NewDiscoveryConfig(nil),
	)

	conns := p2pnode.Connect(node, peers, true)
	for conn := range conns {
		if conn.Error != nil {
			log.Printf("could not connect to %s", conn.Info.ID)
		} else {
			log.Printf("connected to %s", conn.Info.ID)
		}
	}

	return node, nil
}

func main() {
    startNode(nil, nil, []peer.AddrInfo{})
    // ...
}
``` 

