package main

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p"
	inet "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/pkg/browser"
)

//go:embed index.html app.js
var staticFiles embed.FS

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
		opts = append(opts, libp2p.ListenAddrStrings("/ip4/0.0.0.0/tcp/7878/ws"))
	}

	h, err := libp2p.New(opts...)
	if err != nil {
		panic(err)
	}

	var libp2pAddr string
	for _, m := range h.Addrs() {
		addr := fmt.Sprintf("%s/p2p/%s", m, h.ID())
		if strings.Contains(addr, "ws") && strings.Contains(addr, "127.0.0.1") {
			libp2pAddr = addr
		}
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

	sl, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(staticFiles)))

	s := &http.Server{
		Handler: mux,
	}

	go func() {
		s.Serve(sl)
	}()
	defer s.Close()

	addr := "http://" + sl.Addr().String() + "?peer=" + libp2pAddr

	browser.OpenURL(addr)

	select {}
}
