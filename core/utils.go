package core

import (
	"encoding/json"
	"github.com/multiformats/go-multiaddr"
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	logging "github.com/ipfs/go-log/v2"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
)

// PNetSecret creates a new random secret
func PNetSecret() pnet.PSK {
	return randBytes(32)
}

// SerializeAddrInfo marshals (json) the given peer
func SerializeAddrInfo(info peer.AddrInfo) string {
	b, err := json.Marshal(info)
	if err != nil {
		return ""
	}
	return string(b)
}

// SerializePeer marshals (json) the given host's info
func SerializePeer(h host.Host) string {
	pi := peer.AddrInfo{h.ID(), h.Addrs()}
	return SerializeAddrInfo(pi)
}

// MAddrs takes []string and creates the corresponding []multiaddrs
func MAddrs(addrs []string) []multiaddr.Multiaddr {
	maddrs := []multiaddr.Multiaddr{}
	if len(addrs) == 0 {
		return maddrs
	}
	for _, addr := range addrs {
		maddr, _ := multiaddr.NewMultiaddr(addr)
		maddrs = append(maddrs, maddr)
	}
	return maddrs
}

// Peers takes []string ([]multiaddrs) and creates the corresponding peer-info slice
func Peers(addrs []string) []peer.AddrInfo {
	if len(addrs) == 0 {
		return []peer.AddrInfo{}
	}
	maddrs := MAddrs(addrs)
	pis, err := peer.AddrInfosFromP2pAddrs(maddrs...)
	if err != nil {
		log.Fatalf("could not create peers info: %s", err.Error())
	}
	return pis
}

// PrivKey is a utility for managing the peer's private key
func PrivKey(keyPath string) crypto.PrivKey {
	var priv crypto.PrivKey
	if len(keyPath) == 0 { // will not be persisted
		return newPrivKey()
	}
	_, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		priv = newPrivKey()
		go savePrivKey(keyPath, priv)
	} else if err != nil {
		log.Fatal(err)
	} else {
		priv = readPrivKey(keyPath)
	}
	return priv
}

func newPrivKey() crypto.PrivKey {
	pk, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	if err != nil {
		log.Fatal(err)
	}
	return pk
}

func readPrivKey(keyPath string) crypto.PrivKey {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal(err)
	}
	pk, err := crypto.UnmarshalPrivateKey(key)
	if err != nil {
		log.Fatal(err)
	}
	return pk
}

func savePrivKey(keyPath string, pk crypto.PrivKey) {
	data, err := crypto.MarshalPrivateKey(pk)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(keyPath, data, 0400)
	if err != nil {
		log.Fatal(err)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// randBytes creates a random string in the given length
func randBytes(n int) []byte {
	b := make([]byte, n)
	letters := len(letterBytes)
	for i := range b {
		b[i] = letterBytes[rand.Intn(letters)]
	}
	return b[:]
}

func defaultLogger() logging.EventLogger {
	return logging.Logger("libp2p-pnet-node")
}
