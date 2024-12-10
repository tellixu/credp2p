package p2p

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
)

var BootstrapPeers = dht.DefaultBootstrapPeers

const BootstrapConnections = 4

func bootstrapPeerInfo() ([]peer.AddrInfo, error) {
	return peer.AddrInfosFromP2pAddrs(BootstrapPeers...)
}

// ShufflePeerInfos 随机打乱一个地址信息列表
func ShufflePeerInfos(peers []peer.AddrInfo) {
	for i := range peers {
		j := rand.Intn(i + 1)
		peers[i], peers[j] = peers[j], peers[i]
	}
}

func (d *CredHost) Bootstrap() error {
	pisArr, err := bootstrapPeerInfo()
	if err != nil {
		return err
	}
	var pis []peer.AddrInfo

	logger.Info("所有引导节点列表：")
	for _, pi := range pisArr {
		// 本机节点不用连接
		if pi.ID == d.ID() {
			continue
		}
		pis = append(pis, pi)
		d.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
		logger.Info("引导ID：", pi.ID.String())
	}

	//for _, pi := range pis {
	//	if pi.ID == d.ID() {
	//		continue
	//	}
	//	d.host.Peerstore().AddAddrs(pi.ID, pi.Addrs, peerstore.PermanentAddrTTL)
	//}

	count := d.connectBootstrapPeers(pis, BootstrapConnections)
	if count == 0 {
		return fmt.Errorf("failed to connect to bootstrap peers")
	}

	// 保存网络活跃连接
	go d.keepBootstrapConnections(pis)

	if d.dht != nil {
		return d.dht.Bootstrap(d.ctx)
	}

	return nil
}

func (d *CredHost) connectBootstrapPeers(pis []peer.AddrInfo, toconnect int) int {
	count := 0
	// 打乱顺序
	ShufflePeerInfos(pis)

	ctx, cancel := context.WithTimeout(d.ctx, 60*time.Second)
	defer cancel()

	for _, pi := range pis {
		if d.host.Network().Connectedness(pi.ID) == network.Connected {
			continue
		}
		err := d.host.Connect(ctx, pi)
		if err != nil {
			logger.Warn("Error connecting to bootstrap ", "peer=", pi.ID, ",error:", err, ",addrs=", pi.Addrs)
		} else {
			d.host.ConnManager().TagPeer(pi.ID, "bootstrap", 1)
			count++
			toconnect--
			logger.Debug("Connected to bootstrap peer:peerId=", pi.ID)
		}
		if toconnect == 0 {
			break
		}
	}

	return count

}

// 保持最基本的连接数量
func (d *CredHost) keepBootstrapConnections(pis []peer.AddrInfo) {
	ticker := time.NewTicker(15 * time.Minute)
	for {
		<-ticker.C

		conns := d.host.Network().Conns()
		if len(conns) >= BootstrapConnections {
			continue
		}

		toconnect := BootstrapConnections - len(conns)
		d.connectBootstrapPeers(pis, toconnect)
	}
}
