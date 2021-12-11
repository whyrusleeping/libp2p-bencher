package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p"
	inet "github.com/libp2p/go-libp2p-core/network"
	"github.com/libp2p/go-libp2p-core/peer"

	"github.com/dustin/go-humanize"

	cli "github.com/urfave/cli/v2"
)

func main() {

	app := cli.NewApp()

	app.Commands = []*cli.Command{
		clientCmd,
		serverCmd,
	}

	app.RunAndExitOnError()

}

var clientCmd = &cli.Command{
	Name: "client",
	Action: func(cctx *cli.Context) error {
		if !cctx.Args().Present() {
			return fmt.Errorf("must specify multiaddr to connect to")
		}

		opts := []libp2p.Option{
			//libp2p.ConnectionManager(connmgr.NewConnManager(2000, 3000, time.Minute)),
			//libp2p.Identity(peerkey),
			//libp2p.BandwidthReporter(bwc),
			libp2p.DefaultTransports,
			//libp2p.Transport(libp2pquic.NewTransport),
		}

		h, err := libp2p.New(opts...)
		if err != nil {
			return err
		}

		for _, m := range h.Addrs() {
			fmt.Printf("%s/p2p/%s\n", m, h.ID())
		}

		ai, err := peer.AddrInfoFromString(cctx.Args().First())
		if err != nil {
			return err
		}

		if err := h.Connect(context.TODO(), *ai); err != nil {
			return err
		}

		s, err := h.NewStream(context.TODO(), ai.ID, "/bencher")
		if err != nil {
			return err
		}

		start := time.Now()
		n, err := io.Copy(ioutil.Discard, s)
		if err != nil {
			return err
		}

		took := time.Since(start)

		fmt.Printf("read %s bytes in %s. (%s bps)\n", humanize.Bytes(uint64(n)), took, humanize.Bytes(uint64(float64(n)/took.Seconds())))
		return nil
	},
}

var serverCmd = &cli.Command{
	Name: "server",
	Flags: []cli.Flag{
		&cli.Int64Flag{
			Name:  "bytes",
			Usage: "number of bytes to send client",
		},
	},
	Action: func(cctx *cli.Context) error {
		opts := []libp2p.Option{
			//libp2p.ConnectionManager(connmgr.NewConnManager(2000, 3000, time.Minute)),
			//libp2p.Identity(peerkey),
			//libp2p.BandwidthReporter(bwc),
			libp2p.DefaultTransports,
			//libp2p.Transport(libp2pquic.NewTransport),
			libp2p.ListenAddrStrings(
				"/ip4/0.0.0.0/tcp/7878",
				"/ip4/0.0.0.0/tcp/7879/ws",
				"/ip4/0.0.0.0/udp/7880/quic",
			),
		}

		h, err := libp2p.New(opts...)
		if err != nil {
			return nil
		}

		for _, m := range h.Addrs() {
			fmt.Printf("%s/p2p/%s\n", m, h.ID())
		}

		numbytes := cctx.Int64("bytes")

		h.SetStreamHandler("/bencher", func(s inet.Stream) {
			fmt.Println("got a new connection!")
			defer s.Close()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			start := time.Now()

			n, err := io.CopyN(s, r, numbytes)
			if err != nil {
				fmt.Println("copy error: ", err)
			}
			took := time.Since(start)

			fmt.Printf("transfer took %s (%s bps)\n", took, humanize.Bytes(uint64(float64(n)/took.Seconds())))
		})

		select {}
	},
}
