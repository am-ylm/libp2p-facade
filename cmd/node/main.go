package main

import (
	"encoding/json"
	pnet_node "github.com/amirylm/go-libp2p-pnet-node/lib"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func startNode(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (*pnet_node.PrivateNetNode, error) {
	nopts := pnet_node.NewOptions(priv, psk, pnet_node.NewDiscoveryOptions(nil))
	nopts.UseLibp2pOpts = func(opts []libp2p.Option) ([]libp2p.Option, error) {
		return append(opts,
			libp2p.EnableRelay(),
			libp2p.ConnectionManager(connmgr.NewConnManager(10, 50, pnet_node.ConnectionsGrace)),
		), nil
	}
	node, err := pnet_node.NewPrivateNetNode(nopts)
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

	var cfg Config
	p, err := filepath.Abs("./cmd/node/config.json")
	check(err)
	b, err := ioutil.ReadFile(p)
	if err == nil {
		json.Unmarshal(b, &cfg)
	}
	log.Printf("num of peers: %d", len(cfg.Peers))

	rel, err := startNode(psk, priv, cfg.Peers)
	check(err)

	log.Println("node is ready:")
	printHost(rel.Node)

	// wait for a SIGINT or SIGTERM signal
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	<-ch

	errs := rel.Close()
	for _, err := range errs {
		check(err)
	}
}

type Config struct {
	Peers []peer.AddrInfo `json:"peers"`
}

func check(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func printHost(h host.Host) {
	id := h.ID().Pretty()
	log.Printf("%s, listening on:", id)
	for _, addr := range h.Addrs() {
		log.Printf("\t- %s", addr.String())
	}
}