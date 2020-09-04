package main

import (
	"context"
	"encoding/json"
	p2pnode "github.com/amirylm/priv-libp2p-node/core"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func startNode(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (p2pnode.LibP2PNode, error) {
	node := p2pnode.NewBaseNode(context.Background(),
		p2pnode.NewConfig(priv, psk),
		p2pnode.NewDiscoveryConfig(nil),
		libp2p.EnableRelay(),
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
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	psk := []byte("XVlBzgbaiCMRAjWwhTHctcuAxhxKQFDa")

	var cfg p2pnode.Config
	p, err := filepath.Abs("./cmd/node/config.json")
	check(err)
	b, err := ioutil.ReadFile(p)
	if err == nil {
		json.Unmarshal(b, &cfg)
	}
	log.Printf("num of peers: %d", len(cfg.Peers))

	node, err := startNode(psk, priv, cfg.Peers)
	check(err)

	log.Println("node is ready:")
	log.Println(p2pnode.SerializePeer(node.Host()))

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	check(node.Close())
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
