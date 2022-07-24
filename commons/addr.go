package commons

import (
	"fmt"
	"net"
	"time"

	"github.com/libp2p/go-libp2p-core/peer"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/pkg/errors"
)

const (
	checkAddressTimeout = time.Second * 10
)

// CheckAddress checks that some address is accessible and returns error accordingly
func CheckAddress(addr string) error {
	conn, err := net.DialTimeout("tcp", addr, checkAddressTimeout)
	if err != nil {
		return errors.Wrapf(err, "could not dial %s", addr)
	}
	if err := conn.Close(); err != nil {
		return errors.Wrap(err, "could not close connection")
	}
	return nil
}

// BuildMultiAddress creates a multiaddr from the given peer data
func BuildMultiAddress(ipAddr, protocol string, port uint, id peer.ID) (ma.Multiaddr, error) {
	parsedIP := net.ParseIP(ipAddr)
	if parsedIP.To4() == nil && parsedIP.To16() == nil {
		return nil, errors.Errorf("invalid ip address provided: %s", ipAddr)
	}
	maStr := fmt.Sprintf("/ip6/%s/%s/%d", ipAddr, protocol, port)
	if parsedIP.To4() != nil {
		maStr = fmt.Sprintf("/ip4/%s/%s/%d", ipAddr, protocol, port)
	}
	if len(id) > 0 {
		maStr = fmt.Sprintf("%s/p2p/%s", maStr, id.String())
	}
	return ma.NewMultiaddr(maStr)
}
