package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	pnet_node "github.com/amirylm/go-libp2p-pnet-node/lib"
	pnet_relay "github.com/amirylm/go-libp2p-pnet-node/lib/circuit-relay"

	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

func startRelayer(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (*pnet_node.PrivateNetNode, error) {
	rel, err := pnet_relay.NewRelayer(pnet_node.NewOptions(priv, psk))
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

	var cfg pnet_node.Options
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
	log.Println(pnet_node.SerializePeer(rel.Node))

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
