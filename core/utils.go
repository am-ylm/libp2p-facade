package core

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"os"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
)

// PNetSecret creates a new random secret
func PNetSecret() pnet.PSK {
	return randBytes(32)
}

func SerializeAddrInfo(info peer.AddrInfo) string {
	b, err := json.Marshal(info)
	if err != nil {
		return ""
	}
	return string(b)
}

func SerializePeer(h host.Host) string {
	pi := peer.AddrInfo{h.ID(), h.Addrs()}
	return SerializeAddrInfo(pi)
}

func PrivKey(keyPath string) crypto.PrivKey {
	var priv crypto.PrivKey
	if len(keyPath) == 0 { // will not be persisted
		priv, _, err := crypto.GenerateKeyPair(crypto.Ed25519, 1)
		if err != nil {
			log.Fatal(err)
		}
		return priv
	}
	_, err := os.Stat(keyPath)
	if os.IsNotExist(err) {
		priv, _, err = crypto.GenerateKeyPair(crypto.Ed25519, 1)
		if err != nil {
			log.Fatal(err)
		}
		data, err := crypto.MarshalPrivateKey(priv)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(keyPath, data, 0400)
		if err != nil {
			log.Fatal(err)
		}
	} else if err == nil {
		log.Fatal(err)
	} else {
		key, err := ioutil.ReadFile(keyPath)
		if err != nil {
			log.Fatal(err)
		}
		priv, err = crypto.UnmarshalPrivateKey(key)
		if err != nil {
			log.Fatal(err)
		}
	}
	return priv
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
