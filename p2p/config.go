package p2p

import (
	"fmt"
	"github.com/multiformats/go-multiaddr"
	"net"
)

const (
	ReachabilityPublic  = "public"
	ReachabilityPrivate = "private"

	DhtModeServer = "server"
	DhtModeClient = "client"
)

type NetConfigItem struct {
	IP     string `json:"ip" yaml:"ip"`
	Port   int    `json:"port" yaml:"port"`
	PeerID string `json:"peer_id" yaml:"peer_id"`
}

func (that *NetConfigItem) ToMultiaddrs() []multiaddr.Multiaddr {
	addrs := make([]multiaddr.Multiaddr, 0)
	if len(that.PeerID) == 0 {
		return addrs
	}

	ip := net.ParseIP(that.IP)
	if ip == nil {
		// 缺少ip地址，仅使用peerId构建multiaddr.Multiaddr对象
		m1Str := fmt.Sprintf("/p2p/%s", that.PeerID)
		ma, err := multiaddr.NewMultiaddr(m1Str)
		if err != nil {
			return addrs
		}
		addrs = append(addrs, ma)
		return addrs
	} else {
		// 存在ip地址，构建完整的multiaddr.Multiaddr对象
		ipType := "ip4"
		if ip.To4() != nil {
			ipType = "ip4"
		} else if ip.To16() != nil {
			ipType = "ip6"
		}

		m1Str := fmt.Sprintf("/%s/%s/tcp/%d/p2p/%s", ipType, that.IP, that.Port, that.PeerID)
		m2Str := fmt.Sprintf("/%s/%s/udp/%d/quic-v1/p2p/%s", ipType, that.IP, that.Port, that.PeerID)
		//m3Str := fmt.Sprintf("/%s/%s/udp/%d/quic-v1/p2p/%s", ipType, that.IP, that.Port, that.PeerID)
		ma, err := multiaddr.NewMultiaddr(m1Str)
		if err != nil {
			return addrs
		}

		addrs = append(addrs, ma)
		ma, err = multiaddr.NewMultiaddr(m2Str)
		if err != nil {
			return addrs
		}
		addrs = append(addrs, ma)
		return addrs
	}
}

type Config struct {
	//# 网络 reachability：public，private
	//# public：公网可达
	//# private：私网，可能通过nat，中继节点等连接
	Reachability string `yaml:"reachability"`
	// dht模式，默认为client,server,full
	DhtMode           string          `yaml:"dht_mode" json:"dht_mode"`
	Relay             []NetConfigItem `yaml:"relay"` // 中继id
	BootstrapAddr     []NetConfigItem `yaml:"bootstrap_addr"`
	AnnounceAddresses []NetConfigItem `yaml:"announce_addr" json:"announce_addr"`

	P2pPort int `yaml:"p2p_port" json:"p2p_port"`
}

func (that *Config) GetAnnounceAddresses() []multiaddr.Multiaddr {
	addrs := make([]multiaddr.Multiaddr, 0)
	for _, addr := range that.AnnounceAddresses {
		addrs = append(addrs, addr.ToMultiaddrs()...)
	}
	return addrs
}

func (that *Config) GetBootstrapAddr() []multiaddr.Multiaddr {
	addrs := make([]multiaddr.Multiaddr, 0)
	for _, addr := range that.BootstrapAddr {
		addrs = append(addrs, addr.ToMultiaddrs()...)
	}
	return addrs
}
func (that *Config) GetRelayAddr() []multiaddr.Multiaddr {
	addrs := make([]multiaddr.Multiaddr, 0)
	for _, addr := range that.Relay {
		addrs = append(addrs, addr.ToMultiaddrs()...)
	}
	return addrs
}
