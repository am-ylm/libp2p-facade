# go-libp2p-pnet-node

**WIP**

Libp2p private-network node abstraction, it hides a recipe of libp2p's common protocols and concepts (secio/tls, dht, mux, etc...).

The library contains a basic libp2p node, with some pre-defined recipe of protocols and options to launch a private libp2p network.
Libp2p config can be extended with a custom options hook.  
In addition, there is a [circuit-relay](https://docs.libp2p.io/concepts/circuit-relay/) node, which can be extended similarly to the basic node.

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

	pnet_node "github.com/amirylm/go-libp2p-pnet-node/lib"
	
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

