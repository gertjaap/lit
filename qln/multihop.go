package qln

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"github.com/adiabat/bech32"
	"github.com/adiabat/btcutil"
	"github.com/btcsuite/fastsha256"
	"github.com/mit-dci/lit/lnutil"
)

func (nd *LitNode) PayMultihop(dstLNAdr string, coinType uint32, amount int64) (bool, error) {
	var targetAdr [20]byte
	_, adr, err := bech32.Decode(dstLNAdr)
	if err != nil {
		return false, err
	}

	copy(targetAdr[:], adr)
	path, err := nd.FindPath(targetAdr, coinType, amount)

	if err != nil {
		return false, err
	}

	inFlight := new(InFlightMultihop)
	inFlight.Path = path
	nd.InProgMultihop = append(nd.InProgMultihop, inFlight)

	//Connect to the node
	nd.DialPeer(dstLNAdr)
	idx, err := nd.FindPeerIndexByAddress(dstLNAdr)
	if err != nil {
		return false, err
	}

	msg := lnutil.NewMultihopPaymentRequestMsg(idx)
	nd.OmniOut <- msg
	return true, nil
}

func (nd *LitNode) MultihopPaymentRequestHandler(msg lnutil.MultihopPaymentRequestMsg) error {
	// Generate private preimage and send ack with the hash
	fmt.Printf("Received multihop payment request from peer %d", msg.Peer())
	inFlight := new(InFlightMultihop)
	var pkh [20]byte
	id, _ := nd.GetPubHostFromPeerIdx(msg.Peer())
	idHash := fastsha256.Sum256(id[:])
	copy(pkh[:], idHash[:20])
	inFlight.Path = [][20]byte{pkh}
	rand.Read(inFlight.PreImage[:])
	nd.InProgMultihop = append(nd.InProgMultihop, inFlight)

	var hash [20]byte
	copy(hash[:], btcutil.Hash160(inFlight.PreImage[:]))
	outMsg := lnutil.NewMultihopPaymentAckMsg(msg.Peer(), hash)
	nd.OmniOut <- outMsg
	return nil
}

func (nd *LitNode) MultihopPaymentAckHandler(msg lnutil.MultihopPaymentAckMsg) error {
	fmt.Printf("Received multihop payment ack from peer %d, hash %x\n", msg.Peer(), msg.HHash)

	for _, mh := range nd.InProgMultihop {
		targetNode := mh.Path[len(mh.Path)-1]
		targetIdx, _ := nd.FindPeerIndexByAddress(bech32.Encode("ln", targetNode[:]))
		if msg.Peer() == targetIdx {
			fmt.Printf("Found the right pending multihop. Sending setup msg to first hop\n")
			// found the right one. Set this up
			firstHop := mh.Path[1]
			firstHopIdx, _ := nd.FindPeerIndexByAddress(bech32.Encode("ln", firstHop[:]))

			var data [32]byte
			outMsg := lnutil.NewMultihopPaymentSetupMsg(firstHopIdx, mh.Amt, msg.HHash, mh.Path, data)
			fmt.Printf("Sending multihoppaymentsetup to peer %d\n", firstHopIdx)
			nd.OmniOut <- outMsg
		}
	}
	return nil
}

func (nd *LitNode) MultihopPaymentSetupHandler(msg lnutil.MultihopPaymentSetupMsg) error {
	fmt.Printf("Received multihop payment setup from peer %d, hash %x\n", msg.Peer(), msg.HHash)

	found := false
	inFlight := new(InFlightMultihop)

	for _, mh := range nd.InProgMultihop {
		if bytes.Equal(mh.Path[0][:], msg.NodeRoute[0][:]) && mh.Amt == msg.Amount {

			inFlight = mh
			found = true
			// We already know this. If we have a Preimage, then we're the receiving
			// end and we should send a settlement message to the
			// predecessor
			var nullBytes [32]byte
			if mh.PreImage != nullBytes {
				outMsg := lnutil.NewMultihopPaymentSettleMsg(msg.Peer(), mh.PreImage)
				nd.OmniOut <- outMsg
			}
		}
	}

	if !found {
		inFlight.Path = msg.NodeRoute
		inFlight.Amt = msg.Amount
		inFlight.HHash = msg.HHash
		nd.InProgMultihop = append(nd.InProgMultihop, inFlight)
	}

	// Forward
	var pkh [20]byte
	id := nd.IdKey().PubKey().SerializeCompressed()
	idHash := fastsha256.Sum256(id[:])
	copy(pkh[:], idHash[:20])
	var sendToPkh [20]byte
	for i, node := range inFlight.Path {
		if bytes.Equal(pkh[:], node[:]) {
			sendToPkh = inFlight.Path[i+1]
		}
	}

	sendToIdx, _ := nd.FindPeerIndexByAddress(bech32.Encode("ln", sendToPkh[:]))
	msg.PeerIdx = sendToIdx
	nd.OmniOut <- msg
	return nil
}

func (nd *LitNode) MultihopPaymentSettleHandler(msg lnutil.MultihopPaymentSettleMsg) error {
	fmt.Printf("Received multihop payment settle from peer %d\n", msg.Peer())
	found := false
	inFlight := new(InFlightMultihop)

	for _, mh := range nd.InProgMultihop {
		hash := btcutil.Hash160(msg.PreImage[:])
		if bytes.Equal(hash, mh.HHash[:]) {
			inFlight = mh
			found = true
		}
	}

	if !found {
		return fmt.Errorf("Unmatched settle message received")
	}

	// Forward
	var pkh [20]byte
	id := nd.IdKey().PubKey().SerializeCompressed()
	idHash := fastsha256.Sum256(id[:])
	copy(pkh[:], idHash[:20])
	var sendToPkh [20]byte
	for i, node := range inFlight.Path {
		if bytes.Equal(pkh[:], node[:]) {
			sendToPkh = inFlight.Path[i-1]
		}
	}

	sendToIdx, _ := nd.FindPeerIndexByAddress(bech32.Encode("ln", sendToPkh[:]))
	msg.PeerIdx = sendToIdx
	nd.OmniOut <- msg

	return nil
}
