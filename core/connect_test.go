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
	p1 := NewBasePeer(context.Background(), NewConfig(nil, psk, nil))
	defer p1.Close()
	err := p1.DHT().Bootstrap(context.Background())
	assert.Nil(t, err)

	priv2, _, _ := crypto.GenerateKeyPair(crypto.Ed25519, 1)
	p2 := NewBasePeer(context.Background(), NewConfig(priv2, psk, nil))
	defer p2.Close()
	n1Info := peer.AddrInfo{
		ID:    p1.Host().ID(),
		Addrs: p1.Host().Addrs(),
	}
	conns := Connect(p2, []peer.AddrInfo{n1Info}, true)
	for conn := range conns {
		assert.Nil(t, conn.Error)
	}
}
