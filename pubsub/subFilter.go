package pubsub

import (
	"regexp"

	"github.com/libp2p/go-libp2p-core/peer"
	libpubsub "github.com/libp2p/go-libp2p-pubsub"
	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	"github.com/pkg/errors"
)

// NewSubFilter creates a new subscription filter that accepts topics of the given pattern
func NewSubFilter(pattern string, limit int) (libpubsub.SubscriptionFilter, error) {
	reg, err := regexp.Compile(pattern)
	if err != nil {
		return nil, errors.Wrap(err, "could not create sbu filter regexp")
	}
	return &subFilter{reg, limit}, nil
}

type subFilter struct {
	reg   *regexp.Regexp
	limit int
}

// CanSubscribe implements pubsub.SubscriptionFilter
func (sf *subFilter) CanSubscribe(topic string) bool {
	return sf.reg.MatchString(topic)
}

// FilterIncomingSubscriptions implements pubsub.SubscriptionFilter
func (sf *subFilter) FilterIncomingSubscriptions(pi peer.ID, subs []*pb.RPC_SubOpts) ([]*pb.RPC_SubOpts, error) {
	if len(subs) > sf.limit {
		return nil, libpubsub.ErrTooManySubscriptions
	}

	res := libpubsub.FilterSubscriptions(subs, sf.CanSubscribe)

	return res, nil
}
