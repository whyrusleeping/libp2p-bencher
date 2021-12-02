import Websockets from 'libp2p-websockets';
import filters from 'libp2p-websockets/src/filters';
import {Noise} from 'libp2p-noise/dist/src/noise';
import Mplex from 'libp2p-mplex';
import {create} from 'libp2p';
import {Multiaddr} from 'multiaddr';
import PeerId from 'peer-id';

(async () => {
  const libp2p = await create({
    modules: {
      transport: [Websockets],
      connEncryption: [new Noise()],
      streamMuxer: [Mplex],
    },
    config: {
      transport: {
        [Websockets.prototype[Symbol.toStringTag]]: {
          filter: filters.all,
        },
      },
      // do not connect until we dial the protocol
      peerDiscovery: {
        autoDial: false,
      },
    },
  });
  await libp2p.start();

  const peer = new URLSearchParams(window.location.search).get("peer")
  console.log(peer)

  const addr = new Multiaddr(peer);
  const id = addr.getPeerId();
  if (!id) {
    throw new Error('invalid addr');
  }
  const pid = PeerId.createFromB58String(id);
  libp2p.peerStore.addressBook.add(pid, [addr]);

  const t0 = performance.now()
  const {stream} = await libp2p.dialProtocol(pid, '/bencher');
  for await (const _ of stream.source) { }
  const t1 = performance.now()
  console.log('transfer took ', t1-t0) 
})();
