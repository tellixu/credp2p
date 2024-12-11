package main

import (
	"bufio"
	"context"
	"fmt"
	"gitee.com/credata/credp2p/p2p"
	"github.com/ipfs/go-log/v2"
	ps "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"io"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

//# 阿里boot
//- port: 9879
//ip: 47.115.215.250
//peer_id: QmahyM9Sav3A5XeKVpJrfBquKc4m5YNcDze5brVhJiX3Lz
//# 天翼云node1
//- ip: 111.172.230.157
//port: 9879
//peer_id: QmWbA97NBpe8TCw72DtLyxGdZP9FVEJuajiGPXWpgiPsZC
//# 天翼云node2
//- ip: 111.172.230.27
//port: 9879
//peer_id: QmaQPiFw6Uv4PJEQ17JQ5rMQYRLF2J5bE3AULsuM4cDe9N
//# 天翼云node3
//- ip: 111.172.231.214
//port: 9879
//peer_id: QmaDrFaUWRDGVhTbj4gAof28Xx1Qx8eiUEBEHY3MAqYJVV
//# 天翼云node4
//- ip: 111.172.231.11
//port: 9879
//peer_id: Qmc8xtLak9mL3PCZMmzUw4BSAXGyieKe4XxmgAttWNRjKx
//# 天翼云node5
//- ip: 111.172.230.94
//port: 9879
//peer_id: QmShf1MZLQdKCZ9htzbsWNTZLA8QRFqBR3LBZzquSqVJbr

const (
	protocol_x = "cred_p2p/tt/1.0"
)

var logger = log.Logger("cred.sys")

func main() {
	lvl, _ := log.LevelFromString("info")
	//log.SetAllLoggers(lvl)
	log.SetupLogging(log.Config{
		Format: log.ColorizedOutput,
		//Stderr: true,
		Stdout: true,
		//File:   "credLan.log", // todo 输出到文件，需要注意文件的地址判断，否则会会出现错误
		Level: lvl,
		//SubsystemLevels: map[string]log.LogLevel{
		//	"swarm2":  log.LevelWarn,
		//	"relay":   log.LevelWarn,
		//	"connmgr": log.LevelWarn,
		//	"autonat": log.LevelWarn,
		//},
	})
	// 设置环境变量
	//_ = os.Setenv("GOLOG_OUTPUT", "stdout")

	nf := []p2p.NetConfigItem{
		{IP: "47.115.215.250", Port: 9879, PeerID: "QmahyM9Sav3A5XeKVpJrfBquKc4m5YNcDze5brVhJiX3Lz"},
		{IP: "111.172.230.157", Port: 9879, PeerID: "QmWbA97NBpe8TCw72DtLyxGdZP9FVEJuajiGPXWpgiPsZC"},
		{IP: "111.172.230.27", Port: 9879, PeerID: "QmaQPiFw6Uv4PJEQ17JQ5rMQYRLF2J5bE3AULsuM4cDe9N"},
		{IP: "111.172.231.214", Port: 9879, PeerID: "QmaDrFaUWRDGVhTbj4gAof28Xx1Qx8eiUEBEHY3MAqYJVV"},
		{IP: "111.172.231.11", Port: 9879, PeerID: "Qmc8xtLak9mL3PCZMmzUw4BSAXGyieKe4XxmgAttWNRjKx"},
		{IP: "111.172.230.94", Port: 9879, PeerID: "QmShf1MZLQdKCZ9htzbsWNTZLA8QRFqBR3LBZzquSqVJbr"},
	}
	cfg := &p2p.Config{
		Reachability:  p2p.ReachabilityPrivate,
		DhtMode:       p2p.DhtModeClient,
		Relay:         nf,
		BootstrapAddr: nf,
		P2pPort:       4001,
	}

	privateKeyStr, id, err := p2p.GeneratePeerKey()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	fmt.Println("id=", id)

	newCtx, cancel := context.WithCancel(context.Background())

	privateKey, _ := p2p.GetPeerKey(privateKeyStr)
	credHost, err := p2p.NewHost(newCtx, privateKey, cfg)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	credHost.SetHandler(protocol_x, ProtocolTT)
	//---------------
	//go PingTest(credHost)

	pubSubTT(credHost)

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	//go func() {
	//	<-signals
	//	cancel()
	//}()
	q := <-signals
	logger.Infof("received exit signal '%s'", q)
	cancel()
	wg := sync.WaitGroup{}
	wg.Add(1)
	finishedCh := make(chan struct{})
	go func() {
		select {
		case <-time.After(3 * time.Second):
			logger.Fatal("exit timeout reached: terminating")
		case <-finishedCh:
			// ok
			wg.Done()
		case sig := <-signals:
			logger.Fatalf("duplicate exit signal %s: terminating", sig)
		}
	}()
	finishedCh <- struct{}{}
	wg.Wait()
}

func pubSubTT(credHost *p2p.CredHost) {
	topic, err := credHost.GetPubSub().Join("haha_haha")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	//go publishMsg(credHost, topic)

	sub, err := topic.Subscribe()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	go func() {
		processTopic(credHost.GetContext(), sub)
	}()
}

func publishMsg(credHost *p2p.CredHost, topic *ps.Topic) {
	for i := 0; i < 10; i++ {
		time.Sleep(time.Second * 30)
		d := fmt.Sprintf("太棒了:%d", i)
		fmt.Println("准备发送数据:", d)
		err := topic.Publish(credHost.GetContext(), []byte(d))
		if err != nil {
			fmt.Println(err)
		}
	}
}

func processTopic(ctx context.Context, sub *ps.Subscription) {
	for {
		m, err := sub.Next(ctx)
		if err != nil {
			fmt.Println("接收数据失败")
			continue
		}

		go func(msg *ps.Message) {
			fmt.Printf("接收到来自%s的数据:%s\n", m.ReceivedFrom.String(), string(msg.Data))
		}(m)

	}
}

func ProtocolTT(st network.Stream) {
	fmt.Println("assssssssss")
	_, err := io.Copy(st, st)

	if err != nil {
		_ = st.Reset()
	} else {
		_ = st.Close()
	}
}

func PingTest(host *p2p.CredHost) {
	peerIdStr := "12D3KooWPramjK6WTaW9E5MUhJxskbUs9dBHpCownvVzYa2pvLAk"
	peerId, _ := peer.Decode(peerIdStr)

	for i := 0; i < 10; i++ {
		fmt.Printf("begin:%d------------------------------------\n", i)
		pInfo, err := host.FindPeer(context.Background(), peerId)
		if err != nil {
			fmt.Println(err.Error())
			fmt.Println("wait..........")
			time.Sleep(time.Minute)
			continue
		}
		for _, addr := range pInfo.Addrs {
			fmt.Println(addr.String())
		}
		st, err := host.NewStream(context.Background(), peerId, protocol_x)
		if err != nil {
			fmt.Println(err.Error)
			return
		}
		_, _ = st.Write([]byte("hello"))
		r := bufio.NewReader(st)
		buff := make([]byte, 1024)
		l, err := r.Read(buff)
		if err != nil {
			fmt.Println(err.Error)
			return
		}

		fmt.Println(string(buff[:l]))

		//break
		fmt.Println("wait..........")
		time.Sleep(time.Minute)
		continue
	}

}
