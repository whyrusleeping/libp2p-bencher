package main

import (
	"fmt"
	"io"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p"
	inet "github.com/libp2p/go-libp2p-core/network"
)

func main() {

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7878"),
		//libp2p.ConnectionManager(connmgr.NewConnManager(2000, 3000, time.Minute)),
		//libp2p.Identity(peerkey),
		//libp2p.BandwidthReporter(bwc),
		libp2p.DefaultTransports,
		//libp2p.Transport(libp2pquic.NewTransport),
	}
	h, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	for _, m := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", m, h.ID())
	}

	h.SetStreamHandler("/bencher", func(s inet.Stream) {
		r := rand.New(rand.NewSource(time.Now().UnixNano()))
		start := time.Now()

		n, err := io.CopyN(s, r, 100<<20)
		if err != nil {
			fmt.Println("copy error: ", err)
		}
		took := time.Since(start)
		fmt.Printf("transfer took %s (%d bps)\n", took, int(float64(n)/took.Seconds()))
	})

	select {}
}
