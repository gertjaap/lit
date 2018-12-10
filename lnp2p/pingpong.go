package lnp2p

import (
	"fmt"
	"sync"
	"time"

	"github.com/mit-dci/lit/eventbus"
	"github.com/mit-dci/lit/lnwire"
	"github.com/mit-dci/lit/logging"
)

var stopPingMutex sync.Mutex
var stopPing map[*Peer]chan bool

func RegisterPingPongHandlers(mp *MessageProcessor) {
	stopPing = map[*Peer]chan bool{}
	mp.DefineMessage(uint16(lnwire.MsgPing), parsePingMsg, handlePingMsg)
	mp.DefineMessage(uint16(lnwire.MsgPong), parsePongMsg, handlePongMsg)
}

func parsePingMsg(b []byte) (Message, error) {
	return DecodeBoltMsg(&lnwire.Ping{}, b)
}

func parsePongMsg(b []byte) (Message, error) {
	return DecodeBoltMsg(&lnwire.Pong{}, b)
}

func handlePingMsg(p *Peer, m Message) error {
	ping, ok := m.(BoltMsg).InnerMsg.(*lnwire.Ping)
	if !ok {
		return fmt.Errorf("Invalid message type passed")
	}
	reply := make([]byte, ping.NumPongBytes)
	logging.Infof("Received Ping from %d\n", p.GetIdx())
	return p.SendImmediateMessage(BoltMsg{InnerMsg: lnwire.NewPong(reply)})
}

func handlePongMsg(p *Peer, m Message) error {
	_, ok := m.(BoltMsg).InnerMsg.(*lnwire.Pong)
	if !ok {
		return fmt.Errorf("Invalid message type passed")
	}
	logging.Infof("Received pong from %d\n", p.GetIdx())
	return nil
}

func PingPongNewNodeEventHandler() func(eventbus.Event) eventbus.EventHandleResult {
	return func(e eventbus.Event) eventbus.EventHandleResult {
		ee := e.(NewPeerEvent)
		c := make(chan bool, 1)
		stopPingMutex.Lock()
		stopPing[ee.Peer] = c
		stopPingMutex.Unlock()
		pingTicker := time.NewTicker(15 * time.Second)
		go func() {
			for range pingTicker.C {
				select {
				case stop := <-c:
					if stop {
						return
					}
				default:
					logging.Infof("Sending Ping to peer %d\n", ee.Peer.GetIdx())
					// Send an Init message to the new peer
					ping := lnwire.NewPing(32)
					err := ee.Peer.SendImmediateMessage(BoltMsg{InnerMsg: ping})
					if err != nil {
						logging.Errorf("Sending ping message failed: %s\n", err.Error())
					}
				}

			}
		}()

		return eventbus.EHANDLE_OK
	}
}

func PingPongDisconnectEventHandler() func(eventbus.Event) eventbus.EventHandleResult {
	return func(e eventbus.Event) eventbus.EventHandleResult {
		ee := e.(NewPeerEvent)
		stopPingMutex.Lock()
		stopPing[ee.Peer] <- true
		delete(stopPing, ee.Peer)
		stopPingMutex.Unlock()
		return eventbus.EHANDLE_OK
	}
}
