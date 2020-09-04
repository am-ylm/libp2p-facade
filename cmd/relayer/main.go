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

	p2pnode "github.com/amirylm/priv-libp2p-node/core"
	p2prelay "github.com/amirylm/priv-libp2p-node/relay"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

func startRelayer(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) p2pnode.LibP2PNode {
	rel := p2prelay.NewRelayer(context.Background(), p2pnode.NewConfig(priv, psk))

	conns := p2pnode.Connect(rel, peers, true)
	for conn := range conns {
		if conn.Error != nil {
			log.Printf("could not connect to %s", conn.Info.ID)
		} else {
			log.Printf("connected to %s", conn.Info.ID)
		}
	}

	return rel
}

func main() {
	priv, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	psk := []byte("XVlBzgbaiCMRAjWwhTHctcuAxhxKQFDa")

	var cfg p2pnode.Config
	p, err := filepath.Abs("./cmd/relayer/config.json")
	check(err)
	b, err := ioutil.ReadFile(p)
	if err == nil {
		json.Unmarshal(b, &cfg)
	}
	log.Printf("num of peers: %d", len(cfg.Peers))

	rel := startRelayer(psk, priv, cfg.Peers)

	log.Println("circuit relay node is ready:")
	log.Println(p2pnode.SerializePeer(rel.Host()))

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	check(rel.Close())
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}
