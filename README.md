# libp2p-facade

**WIP**

This module consists of libp2p peer abstraction, utilities, configurations and setup. 
The idea is to encapsulate libp2p's common components (pubsub, dht, ipld, etc...) using the same API cross-project.

Inspired by [ipfs-lite](https://github.com/hsanjuan/ipfs-lite), which is an alternative to a full IPFS, 
and comes by default with IPLD (DAGService).

This package makes it easy to configure several types of libp2p peers:
- LibP2PPeer (`./core/node.go`) is the most minimal interface
    - local datastore, DHT, pubsub
    - a [circuit-relay](https://docs.libp2p.io/concepts/circuit-relay/) peer for example, doesn't need more than that (see `./core/circuit-relay`)
- StoragePeer (`./storage/peer.go`) - a peer with distributed storage for IPLD data
    - similar to `ipfs-lite`
    - [ipld.DAGService](https://godoc.org/github.com/ipfs/go-ipld-format#DAGService)
    - [crdt](https://github.com/ipfs/go-ds-crdt) 
    could be configured easily to provide state consistency, see `./storage/crdt.go`

## Install

As a library:

```bash
go get github.com/amirylm/libp2p-facade
```

## Usage

More examples available in:
  - tests in this project
  - [amir-yahalom/go-csn](https://github.com/amir-yahalom/go-csn) > `./cmd` folder

```go
package main

import (
    "log"

	p2pfacade "github.com/amirylm/libp2p-facade/core"
	
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

func startPeer(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (p2pfacade.LibP2PPeer, error) {
    p := p2pfacade.NewBasePeer(context.Background(),
		p2pfacade.NewConfig(priv, psk, nil),
		p2pfacade.NewDiscoveryConfig(nil),
	)

	conns := p2pfacade.Connect(p, peers, true)
	for conn := range conns {
		if conn.Error != nil {
			log.Printf("could not connect to %s", conn.Info.ID)
		} else {
			log.Printf("connected to %s", conn.Info.ID)
		}
	}

	return p, nil
}

func main() {
    startPeer(nil, nil, []peer.AddrInfo{})
    // ...
}
``` 

