package p2p

import (
	dht "github.com/libp2p/go-libp2p-kad-dht"
)

type HashDht struct {
	*dht.IpfsDHT
}

//func (d *CredHost) FindPeer(ctx context.Context, peerId peer.ID) (peer.AddrInfo, error) {
//	return FindPeer(ctx, peerId)
//}
