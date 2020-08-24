# go-libp2p-pnet-node

**WIP**

libp2p private-network node

## Install

As a library:

```bash
go get github.com/amirylm/go-libp2p-pnet-node
```


## Usage

See `./cmd` folder and tests for more concrete examples.

```go
package main

import (
    "log"

	pnet_node "github.com/amirylm/go-libp2p-pnet-node"
	
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

func startNode(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (*pnet_node.PrivateNetNode, error) {
	nopts := pnet_node.NewOptions(priv, psk)
	nopts.Libp2pOpts = func() ([]libp2p.Option, error) {
		return []libp2p.Option{
			libp2p.EnableRelay(),
			libp2p.ConnectionManager(connmgr.NewConnManager(10, 50, pnet_node.ConnectionsGrace)),
		}, nil
	}
	node, _ := pnet_node.NewPrivateNetNode(nopts)

	conns := node.ConnectToPeers(peers, true)
	for conn := range conns {
		if conn.Error != nil {
			log.Printf("could not connect to %s", conn.ID)
		} else {
			log.Printf("connected to %s", conn.ID)
		}
	}

	return node, nil
}

func main() {
    startNode(nil, nil, []peer.AddrInfo{})
    // ...
}
``` 

