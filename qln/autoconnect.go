package qln

import (
	"bytes"
	"fmt"
	"time"

	"github.com/adiabat/bech32"
	"github.com/btcsuite/fastsha256"
)

// AutoReconnect will start listening for incoming connections
// and attempt to automatically reconnect to all
// previously known peers.
func (nd *LitNode) AutoReconnect(listenPort string, interval int64) {
	// Listen myself
	nd.TCPListener(listenPort)

	// Reconnect to other nodes after a timeout
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	go func() {
		for range ticker.C {
			fmt.Println("Reconnecting to known peers")
			var empty [33]byte
			i := uint32(0)
			for {
				pubKey, _ := nd.GetPubHostFromPeerIdx(i)
				if pubKey == empty {
					fmt.Printf("Done, tried %d hosts\n", i)
					break
				}
				i++
				alreadyConnected := false

				nd.RemoteMtx.Lock()
				for _, con := range nd.RemoteCons {
					if bytes.Equal(con.Con.RemotePub.SerializeCompressed(), pubKey[:]) {
						alreadyConnected = true
					}
				}
				nd.RemoteMtx.Unlock()

				if alreadyConnected {
					continue
				}

				idHash := fastsha256.Sum256(pubKey[:])
				adr := bech32.Encode("ln", idHash[:20])

				err := nd.DialPeer(adr)

				if err != nil {
					fmt.Printf("Could not restore connection to %s: %s\n", adr, err.Error())
				}
			}
		}
	}()

}
