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

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

func startRelayer(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (*p2pnode.PrivateNetNode, error) {
	rel, err := p2pnode.NewRelayer(context.Background(), p2pnode.NewOptions(priv, psk, nil))
	if err != nil {
		return rel, err
	}

	conns := rel.ConnectToPeers(peers, true)
	for conn := range conns {
		if conn.Error != nil {
			log.Printf("could not connect to %s", conn.ID)
		} else {
			log.Printf("connected to %s", conn.ID)
		}
	}

	return rel, err
}

func main() {
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	psk := []byte("XVlBzgbaiCMRAjWwhTHctcuAxhxKQFDa")

	var cfg p2pnode.Options
	p, err := filepath.Abs("./cmd/relayer/config.json")
	check(err)
	b, err := ioutil.ReadFile(p)
	if err == nil {
		json.Unmarshal(b, &cfg)
	}
	log.Printf("num of peers: %d", len(cfg.Peers))

	rel, err := startRelayer(psk, priv, cfg.Peers)
	check(err)

	log.Println("circuit relay node is ready:")
	log.Println(p2pnode.SerializePeer(rel.Node))

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	errs := rel.Close()
	for _, err := range errs {
		check(err)
	}
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
