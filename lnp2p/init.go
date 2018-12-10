package lnp2p

import (
	"fmt"

	"github.com/mit-dci/lit/eventbus"
	"github.com/mit-dci/lit/lnwire"
	"github.com/mit-dci/lit/logging"
)

func RegisterInitHandler(mp *MessageProcessor) {
	mp.DefineMessage(uint16(lnwire.MsgInit), parseInitMsg, handleInitMsg)
}

func parseInitMsg(b []byte) (Message, error) {
	return DecodeBoltMsg(&lnwire.Init{}, b)
}

func handleInitMsg(p *Peer, m Message) error {
	init, ok := m.(BoltMsg).InnerMsg.(*lnwire.Init)
	if !ok {
		return fmt.Errorf("Invalid message type passed")
	}

	logging.Infof("Received BOLT Init message: %x\n", init.GlobalFeatures)

	return nil
}

func InitNewNodeEventHandler() func(eventbus.Event) eventbus.EventHandleResult {
	return func(e eventbus.Event) eventbus.EventHandleResult {
		ee := e.(NewPeerEvent)

		logging.Infof("Sending Init message to new peer %d\n", ee.Peer.GetIdx())
		// Send an Init message to the new peer
		init := lnwire.NewInitMessage(lnwire.NewRawFeatureVector(), lnwire.NewRawFeatureVector())
		var initMsg BoltMsg
		initMsg.InnerMsg = init
		err := ee.Peer.SendImmediateMessage(initMsg)
		if err != nil {
			logging.Errorf("Sending init message failed: %s\n", err.Error())
		}

		return eventbus.EHANDLE_OK
	}
}
