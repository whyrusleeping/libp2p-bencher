package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"time"

	"github.com/libp2p/go-libp2p"
	inet "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
)

func main() {

	listener := len(os.Args) == 1

	opts := []libp2p.Option{
		//libp2p.ConnectionManager(connmgr.NewConnManager(2000, 3000, time.Minute)),
		//libp2p.Identity(peerkey),
		//libp2p.BandwidthReporter(bwc),
		libp2p.DefaultTransports,
		//libp2p.Transport(libp2pquic.NewTransport),
	}

	if listener {
		opts = append(opts, libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7878"))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	for _, m := range h.Addrs() {
		fmt.Printf("%s/p2p/%s\n", m, h.ID())
	}

	if !listener {
		ai, err := peer.AddrInfoFromString(os.Args[1])
		if err != nil {
			panic(err)
		}

		if err := h.Connect(context.TODO(), *ai); err != nil {
			panic(err)
		}

		s, err := h.NewStream(context.TODO(), ai.ID, "/bencher")
		if err != nil {
			panic(err)
		}

		start := time.Now()
		n, err := io.Copy(ioutil.Discard, s)
		if err != nil {
			panic(err)
		}

		took := time.Since(start)

		fmt.Printf("read %d bytes in %s. (%d bps)\n", n, took, int(float64(n)/took.Seconds()))
		return
	}

	h.SetStreamHandler("/bencher", func(s inet.Stream) {
		defer s.Close()
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
