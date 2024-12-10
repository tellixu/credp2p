package p2p

import (
	"context"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/peer"
)

type HashDht struct {
	*dht.IpfsDHT
}

func (d *HashDht) FindPeer(ctx context.Context, peerId peer.ID) (peer.AddrInfo, error) {
	return d.FindPeer(ctx, peerId)
}
