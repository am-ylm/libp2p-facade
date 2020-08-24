package node

import (
	"encoding/json"
	pnet_node "github.com/amirylm/go-libp2p-pnet-node"
	"github.com/libp2p/go-libp2p"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func startNode(psk pnet.PSK, priv crypto.PrivKey, peers []peer.AddrInfo) (*pnet_node.PrivateNetNode, error) {
	nopts := pnet_node.NewOptions(priv, psk)
	nopts.Libp2pOpts = func() ([]libp2p.Option, error) {
		return []libp2p.Option{
			libp2p.EnableRelay(),
			libp2p.ConnectionManager(connmgr.NewConnManager(10, 50, pnet_node.ConnectionsGrace)),
		}, nil
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

	peers := []peer.AddrInfo{}
	b, err := ioutil.ReadFile("./peers.json")
	if err == nil {
		json.Unmarshal(b, peers)
	}

	rel, err := startNode(psk, priv, peers)
	check(err)

	log.Println("ready...")

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