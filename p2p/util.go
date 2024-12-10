package p2p

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

func GetPeerAddrInfos(mAddr []multiaddr.Multiaddr) []peer.AddrInfo {
	ds := make([]peer.AddrInfo, 0, len(mAddr))

	for i := range mAddr {
		info, err := peer.AddrInfoFromP2pAddr(mAddr[i])
		if err != nil {
			logger.Errorw("failed to convert bootstrapper address to peer addr info", "address",
				mAddr[i].String(), err, "err")
			continue
		}
		ds = append(ds, *info)
	}
	return ds
}

func ToMAddr(addrs []string) ([]multiaddr.Multiaddr, error) {
	var multiAddrs []multiaddr.Multiaddr
	for _, addr := range addrs {
		m2, e2 := multiaddr.NewMultiaddr(addr)
		if e2 != nil {
			logger.Warn(e2.Error())
			return nil, e2
		} else {
			multiAddrs = append(multiAddrs, m2)
		}
	}
	return multiAddrs, nil
}

func Md5(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}
