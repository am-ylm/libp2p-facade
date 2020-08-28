package lib

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"log"
	"sync"
	"testing"
	"time"
)

func TestPubSubEmitter(t *testing.T) {
	n := 4
	psk := PNetSecret()
	nodes, err := setupNodesGroup(n, psk)
	assert.Nil(t, err)
	assert.Equal(t, n, len(nodes))
	log.Println("after discovery")

	var pswg sync.WaitGroup
	data := []byte("data:my-topic")
	pswg.Add(1)
	sub1, err := nodes[0].Emitter.Subscribe("my-topic")
	assert.Nil(t, err)
	go func() {
		for {
			msg, err := sub1.Next(nodes[0].ctx)
			assert.Nil(t, err)
			assert.True(t, bytes.Equal(data, msg.Data))
			pswg.Done()
			return
		}
	}()

	sub2, err := nodes[1].Emitter.Subscribe("other-topic")
	assert.Nil(t, err)
	go func() {
		for {
			sub2.Next(nodes[1].ctx)
			assert.Fail(t, "should not receive a message")
			return
		}
	}()

	topic3, err := nodes[2].Emitter.Topic("my-topic")
	assert.Nil(t, err)
	go func() {
		time.Sleep(1000 * time.Millisecond)
		topic3.Publish(nodes[2].ctx, data[:])
	}()
	pswg.Wait()
}
