package core

import (
	"context"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewNodeConnect(t *testing.T) {
	psk := PNetSecret()
	n1 := NewBaseNode(context.Background(), NewConfig(nil, psk, nil), nil)
	err := n1.DHT().Bootstrap(context.Background())
	assert.Nil(t, err)

	priv2, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	n2 := NewBaseNode(context.Background(), NewConfig(priv2, psk, nil), nil)
	n1Info := peer.AddrInfo{
		ID:    n1.Host().ID(),
		Addrs: n1.Host().Addrs(),
	}
	conns := Connect(n2, []peer.AddrInfo{n1Info}, true)
	for conn := range conns {
		assert.Nil(t, conn.Error)
	}
}
