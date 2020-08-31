package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	p2pnode "github.com/amirylm/priv-libp2p-node/lib"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

func startNode(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (*p2pnode.PrivateNetNode, error) {
	nopts := p2pnode.NewOptions(priv, psk, p2pnode.NewDiscoveryOptions(nil))
	nopts.UseLibp2pOpts = func(opts []libp2p.Option) ([]libp2p.Option, error) {
		return append(opts,
			libp2p.EnableRelay(),
			libp2p.ConnectionManager(connmgr.NewConnManager(10, 50, p2pnode.ConnectionsGrace)),
		), nil
	}
	node, err := p2pnode.NewPrivateNetNode(context.Background(), nopts)
	check(err)

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
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	psk := []byte("XVlBzgbaiCMRAjWwhTHctcuAxhxKQFDa")

	var cfg p2pnode.Options
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
	log.Println(p2pnode.SerializePeer(node.Node))

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	errs := node.Close()
	for _, err := range errs {
		check(err)
	}
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
