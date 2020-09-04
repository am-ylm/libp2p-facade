package core

import (
	"encoding/json"
	"math/rand"

	logging "github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/libp2p/go-libp2p-core/pnet"
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
