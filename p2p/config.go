package p2p

import (
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

const (
	ReachabilityPublic  = "public"
	ReachabilityPrivate = "private"

	DhtModeServer = "server"
	DhtModeClient = "client"
)

type MulAddrArr []string

func (t MulAddrArr) ToMultiaddr() ([]multiaddr.Multiaddr, error) {

	addrs := make([]multiaddr.Multiaddr, 0)
	for _, addr := range t {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, ma)
	}

	return addrs, nil
}

func (t MulAddrArr) ToAddrInfos() ([]peer.AddrInfo, error) {

	addrs := make([]multiaddr.Multiaddr, 0)
	for _, addr := range t {
		ma, err := multiaddr.NewMultiaddr(addr)
		if err != nil {
			return nil, err
		}
		addrs = append(addrs, ma)
	}

	return peer.AddrInfosFromP2pAddrs(addrs...)
}

type ACLConfig struct {
	//# `allow_peers` is a white list of allowed client side peers to access
	//# default to empty, that means allow all.
	AllowPeers []string `json:"allow_peers" yaml:"allow_peers"`
	//# `allow_subnets` is a white list of allowed subnets that client side peers to access
	//# for example: ["127.0.0.1/32", "::1/128"], that means only allowing local peers.
	AllowSubnets []string `json:"allow_subnets" yaml:"allow_subnets"`
}

type Config struct {
	// dht模式，默认为client,server,full
	DhtMode           string     `yaml:"dht_mode" json:"dht_mode"`
	Relay             MulAddrArr `yaml:"relay" json:"relay"` // 中继id
	BootstrapAddr     MulAddrArr `yaml:"bootstrap_addr" json:"bootstrap_addr"`
	AnnounceAddresses MulAddrArr `yaml:"announce_addr" json:"announce_addr"`
	Listener          MulAddrArr `json:"listener" yaml:"listener"`
	ACL               ACLConfig  `json:"acl" yaml:"acl"` // 控制系统
}

//
//func (that *Config) GetAnnounceAddresses() []multiaddr.Multiaddr {
//	addrs := make([]multiaddr.Multiaddr, 0)
//	for _, addr := range that.AnnounceAddresses {
//		addrs = append(addrs, addr.ToMultiaddrs()...)
//	}
//	return addrs
//}
//
//func (that *Config) GetBootstrapAddr() []multiaddr.Multiaddr {
//	addrs := make([]multiaddr.Multiaddr, 0)
//	for _, addr := range that.BootstrapAddr {
//		addrs = append(addrs, addr.ToMultiaddrs()...)
//	}
//	return addrs
//}
//func (that *Config) GetRelayAddr() []multiaddr.Multiaddr {
//	addrs := make([]multiaddr.Multiaddr, 0)
//	for _, addr := range that.Relay {
//		addrs = append(addrs, addr.ToMultiaddrs()...)
//	}
//	return addrs
//}
