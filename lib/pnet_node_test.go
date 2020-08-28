package lib

import (
	"context"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPrivateNetNode(t *testing.T) {
	psk := PNetSecret()
	n1, err := NewPrivateNetNode(context.Background(), NewOptions(nil, psk, nil))
	assert.Nil(t, err)
	err = n1.Dht.Bootstrap(context.Background())
	assert.Nil(t, err)

	priv2, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	n2, err := NewPrivateNetNode(context.Background(), NewOptions(priv2, psk, nil))
	assert.Nil(t, err)
	n1Info := peer.AddrInfo{
		ID:    n1.Node.ID(),
		Addrs: n1.Node.Addrs(),
	}
	conns := n2.ConnectToPeers([]peer.AddrInfo{n1Info}, true)
	for conn := range conns {
		assert.Nil(t, conn.Error)
	}
}