package lnp2p

import (
	"github.com/mit-dci/lit/eventbus"
	"github.com/mit-dci/lit/lncore"
	"github.com/mit-dci/lit/lndc"
	"github.com/mit-dci/lit/logging"
)

type listeningthread struct {
	listener *lndc.Listener
}

func acceptConnections(listener *lndc.Listener, port int, pm *PeerManager) {

	// Set this up in-advance.
	stopEvent := &StopListeningPortEvent{
		Port:   port,
		Reason: "panic",
	}

	// Do this now in case we panic so we can do cleanup.
	defer publishStopEvent(stopEvent, pm.ebus)

	// Actually start listening for connections.
	for {

		logging.Infof("Waiting to accept a new connection...\n")

		netConn, err := listener.Accept()
		if err != nil {
			if err.Error() == "lndc connection closed" {
				logging.Infof("LNDC connection closed, exiting\n")
				break // usually means the socket was closed
			} else {
				logging.Debugf("got an error while accepting connection: %s\n", err.Error())
				continue // If the transport isn't closed we should keep accepting connections
			}
		}

		lndcConn, ok := netConn.(*lndc.Conn)
		if !ok {
			// this should never happen
			logging.Errorf("didn't get an lndc connection from listener, quitting!\n")
			netConn.Close()
			continue
		}

		remotePubkey := pubkey(lndcConn.RemotePub())
		remoteLitAddr := convertPubkeyToLitAddr(remotePubkey)
		remoteNetAddr := lndcConn.RemoteAddr().String()

		logging.Infof("New connection from %s at %s\n", remoteLitAddr, remoteNetAddr)

		// Create the actual peer object since we need to interface with db to do this
		// TODO: remove this
		newPeer := &Peer{
			lnaddr:   remoteLitAddr,
			nickname: nil,
			conn:     lndcConn,
			idpubkey: remotePubkey,
			idx:      nil,
		}

		// Read the peer info from the DB.
		pi, err := pm.peerdb.GetPeerInfo(remoteLitAddr) // search based on remote peer address
		if err != nil {
			logging.Warnf("problem loading peer info in DB: %s\n", err.Error())
			netConn.Close()
			continue
		}

		if pi != nil {
			// we already have this peer in the db, but we want to udpate this
			// the update method is a bit weird, so we delete it directly and
			// replacing it with a new one.
			// TODO: Fix the UpdatePeer method to behave as expected
			logging.Info("Found peer in peerdb. Don't disconnect, just delete", pi.PeerIdx)
			newPeer.idx = &pi.PeerIdx
			pm.peerdb.DeletePeer(remoteLitAddr)
		} else {
			temp, err := pm.peerdb.GetUniquePeerIdx()
			if err != nil {
				logging.Errorf("problem getting unique peeridx: %s\n", err.Error())
			}
			newPeer.idx = &temp
		}

		pi = &lncore.PeerInfo{
			LnAddr:   &remoteLitAddr,
			Nickname: nil,
			NetAddr:  &remoteNetAddr,
			PeerIdx:  *newPeer.idx,
		}
		err = pm.peerdb.AddPeer(remoteLitAddr, *pi)
		if err != nil {
			// don't close it, I guess
			logging.Errorf("problem saving peer info to DB: %s\n", err.Error())
		}

		// Don't do any locking here since registerPeer takes a lock and Go's
		// mutex isn't reentrant.
		pm.registerPeer(newPeer)

		// Start a goroutine to process inbound traffic for this peer.
		go processConnectionInboundTraffic(newPeer, pm)
	}

	// Update the stop reason.
	stopEvent.Reason = "closed"

	// Then delete the entry from listening ports.
	pm.mtx.Lock()
	delete(pm.listeningPorts, port)
	pm.mtx.Unlock()

	// after this the stop event will be published
	logging.Infof("Stopped listening on %s\n", port)
}

func processConnectionInboundTraffic(peer *Peer, pm *PeerManager) {

	// Set this up in-advance.
	dcEvent := &PeerDisconnectEvent{
		Peer:   peer,
		Reason: "panic",
	}

	// Do this now in case we panic so we can do cleanup.
	defer publishDisconnectEvent(dcEvent, pm.ebus)

	// TODO Have chanmgr deal with channels after peer connection brought up. (eventbus)

	for {

		// Make a buf and read into it.
		buf := make([]byte, 1<<24)
		n, err := peer.conn.Read(buf)
		if err != nil {
			logging.Warnf("Error reading from peer: %s\n", err.Error())
			peer.conn.Close()
			return
		}

		logging.Debugf("Got message of len %d from peer %s\n", n, peer.GetLnAddr())

		// Send to the message processor.
		err = pm.mproc.HandleMessage(peer, buf[:n])
		if err != nil {
			logging.Errorf("Error proccessing message: %s\n", err.Error())
		}

	}

}

func publishStopEvent(event *StopListeningPortEvent, ebus *eventbus.EventBus) {
	ebus.Publish(*event)
}

func publishDisconnectEvent(event *PeerDisconnectEvent, ebus *eventbus.EventBus) {
	ebus.Publish(*event)
}
