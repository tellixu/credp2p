package p2p

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/ipfs/go-log/v2"
	"github.com/libp2p/go-libp2p"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/host/autorelay"
	"github.com/libp2p/go-libp2p/p2p/protocol/ping"
	tls "github.com/libp2p/go-libp2p/p2p/security/tls"
	"github.com/multiformats/go-multiaddr"
	"github.com/tellixu/credp2p/p2p/discovery"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var logger = log.Logger("cred.p2p")

var (
	serviceName = "cred_p2p/1.0"
)

func SetServiceName(sn string) {
	serviceName = sn
}

func GetServiceName() string {
	return serviceName
}

type CredHost struct {
	ctx    context.Context
	cancel context.CancelFunc
	host   host.Host

	dht *HashDht
	//pubsub *ps.PubSub

	mx sync.Mutex
	// closed is set when the daemon is shutting down
	closed bool

	//privKey  crypto.PrivKey // 私钥
	CryptoKey *CryptoKeyInfo

	cfg *Config
}

func NewHost(ctx context.Context, privateKey crypto.PrivKey, cfg *Config) (*CredHost, error) {
	if len(cfg.BootstrapAddr) == 0 {
		logger.Warn("BootstrapAddr is empty")
		return nil, fmt.Errorf("BootstrapAddr is empty")
	}
	bootPeers, err := cfg.BootstrapAddr.ToMultiaddr()
	if err != nil {
		logger.Warn("BootstrapAddr is fail")
		return nil, err
	}
	BootstrapPeers = bootPeers

	ctx, cancel := context.WithCancel(ctx)
	credHost := &CredHost{
		ctx:       ctx,
		cancel:    cancel,
		cfg:       cfg,
		CryptoKey: NewCryptoKeyInfo(privateKey),
	}

	//listenerAddStrs := []string{
	//	fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", cfg.P2pPort),         // tcp连接
	//	fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic-v1", cfg.P2pPort), // 用于QUIC传输的UDP端点
	//	fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", cfg.P2pPort),      // websocket
	//
	//	// :: 格式的ipv6，支持所有的ipv4和ipv6的访问
	//	fmt.Sprintf("/ip6/::/tcp/%d", cfg.P2pPort),         // tcp连接
	//	fmt.Sprintf("/ip6/::/udp/%d/quic-v1", cfg.P2pPort), // 用于QUIC传输的UDP端点
	//	fmt.Sprintf("/ip6/::/tcp/%d/ws", cfg.P2pPort),      // websocket
	//}

	var opts []libp2p.Option = []libp2p.Option{
		libp2p.Identity(privateKey),
		libp2p.UserAgent(serviceName),
		//libp2p.ListenAddrs(listenerAddr...),
		libp2p.Security(tls.ID, tls.New),
		libp2p.WithDialTimeout(time.Second * 60),
		libp2p.EnableHolePunching(),
		libp2p.EnableRelay(),
		libp2p.DefaultTransports,
	}

	if len(cfg.Listener) == 0 {
		opts = append(opts, libp2p.NoListenAddrs)
	} else {
		listenerAddr, err := cfg.Listener.ToMultiaddr()
		if err != nil {
			logger.Warn(err.Error())
			return nil, err
		}
		opts = append(opts, libp2p.ListenAddrs(listenerAddr...))

		opts = append(opts, libp2p.NATPortMap(), libp2p.EnableNATService(), libp2p.EnableAutoNATv2())

		opts = append(opts,
			libp2p.EnableRelayService())

		opts = append(opts,
			libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
				dhtBootPeer := GetPeerAddrInfos(BootstrapPeers)

				dhtopts := []dht.Option{
					//dht.Mode(dht.ModeClient),
					dht.BootstrapPeers(dhtBootPeer...),
				}

				if cfg.DhtMode == DhtModeServer {
					dhtopts = append(dhtopts, dht.Mode(dht.ModeServer))
				} else if cfg.DhtMode == DhtModeClient {
					dhtopts = append(dhtopts, dht.Mode(dht.ModeClient))
				}

				idht, err := dht.New(ctx, h, dhtopts...)
				//if err == nil {
				//	idht.Bootstrap(rootCtx)
				//}
				credHost.dht = &HashDht{IpfsDHT: idht}
				return idht, err
			}),
		)

		if len(cfg.AnnounceAddresses) > 0 {
			// 定义为public可达
			opts = append(opts, libp2p.ForceReachabilityPublic())
			announceAddr, err := cfg.AnnounceAddresses.ToMultiaddr()
			if err != nil {
				logger.Warn(err.Error())
				return nil, err
			}
			opts = append(opts, libp2p.AddrsFactory(func([]multiaddr.Multiaddr) []multiaddr.Multiaddr {
				return announceAddr
			}))
		} else {
			if len(cfg.Relay) > 0 {
				relayAddr, err := cfg.Relay.ToAddrInfos()
				if err != nil {
					logger.Warn(err.Error())
					return nil, err
				}
				//relayAddrInfos := GetPeerAddrInfos(cfg.GetRelayAddr())
				opts = append(opts, libp2p.EnableAutoRelayWithStaticRelays(
					relayAddr,
					autorelay.WithBootDelay(20*time.Second),
				))
			}
		}
	}
	// 添加访问控制
	acl, err := NewACL(&cfg.ACL)
	if err != nil {
		logger.Error(err)
		return nil, err
	}
	opts = append(opts, libp2p.ConnectionGater(acl))

	h, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}
	credHost.host = h

	if len(cfg.Listener) > 0 {
		err = credHost.Bootstrap()
		if err != nil {
			logger.Warn(err.Error())
			return nil, err
		}

		// mDns discovery
		m := &discovery.MDNS{}
		m.DiscoveryServiceTag = Md5([]byte(serviceName))
		_ = m.Run(ctx, h)

		//pubSubOption:=[]ps.Option{
		//	ps.with
		//}
		// 允许使用limited连接，中继时会用到
		//pubsub, err := ps.NewGossipSub(network.WithAllowLimitedConn(ctx, "credata_p2p_pubsub"), h, ps.WithDiscovery(drouting.NewRoutingDiscovery(credHost.dht)))
		//pubsub, err := ps.NewGossipSub(network.WithAllowLimitedConn(ctx, "credata_p2p_pubsub"), h)
		//drouting "github.com/libp2p/go-libp2p/p2p/discovery/routing"
		//pubsub, err := ps.NewGossipSub(network.WithAllowLimitedConn(ctx, "credata_p2p_pubsub"), h)
		//if err != nil {
		//	logger.Warn(err.Error())
		//	return nil, err
		//}
		//credHost.pubsub = pubsub
		ping.NewPingService(h)
	}
	logger.Info("peer id is:", h.ID().String())
	return credHost, nil
}

func (d *CredHost) ID() peer.ID {
	return d.host.ID()
}

func (d *CredHost) GetHost() host.Host {
	return d.host
}

func (d *CredHost) GetContext() context.Context {
	return d.ctx
}
func (d *CredHost) isClosed() bool {
	d.mx.Lock()
	defer d.mx.Unlock()
	return d.closed
}

func (d *CredHost) Close() error {
	d.mx.Lock()
	d.closed = true
	d.mx.Unlock()

	d.cancel()

	var merr *multierror.Error
	if err := d.host.Close(); err != nil {
		merr = multierror.Append(err)
	}
	if merr != nil {
		return merr.ErrorOrNil()
	}
	return nil
}

func (d *CredHost) IsConnected(peerID peer.ID) bool {
	return d.host.Network().Connectedness(peerID) != network.NotConnected
}

func (d *CredHost) FindPeer(ctx context.Context, peerId peer.ID) (peer.AddrInfo, error) {
	peerStore := d.host.Peerstore()
	peerAddrInfo := peerStore.PeerInfo(peerId)
	if len(peerAddrInfo.Addrs) == 0 {
		return d.dht.FindPeer(ctx, peerId)
	}
	return peerAddrInfo, nil
}

// NewStream 创建一个流，如果创建失败，则重新从dht中获取，再进行连接
func (d *CredHost) NewStream(ctx context.Context, id peer.ID, proto protocol.ID) (network.Stream, error) {
	if !d.IsConnected(id) {
		distAddrInfo, e1 := d.dht.FindPeer(ctx, id)
		if e1 != nil {
			return nil, e1
		}
		e1 = d.host.Connect(ctx, distAddrInfo)
		if e1 != nil {
			return nil, e1
		}
	}

	// 允许使用limited连接，中继时会用到
	ctx = network.WithAllowLimitedConn(ctx, "credata_p2p")
	return d.host.NewStream(ctx, id, proto)
}

//func (d *CredHost) GetPubSub() *ps.PubSub {
//	return d.pubsub
//}

func (d *CredHost) SetHandler(pid protocol.ID, handler network.StreamHandler) {
	d.host.SetStreamHandler(pid, handler)
}
func (d *CredHost) RemoveHandler(pid protocol.ID) {
	d.host.RemoveStreamHandler(pid)
}

func ContextWithSignal(ctx context.Context) context.Context {
	newCtx, cancel := context.WithCancel(ctx)
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	go func() {
		<-signals
		cancel()
	}()
	return newCtx
}
