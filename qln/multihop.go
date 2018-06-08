package qln

import (
	"crypto/rand"

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
	for _, mh := range nd.InProgMultihop {
		targetNode := mh.Path[len(mh.Path)-1]
		targetIdx, _ := nd.FindPeerIndexByAddress(bech32.Encode("ln", targetNode[:]))
		if msg.Peer() == targetIdx {
			// found the right one. Set this up
			firstHop := mh.Path[1]
			firstHopIdx, _ := nd.FindPeerIndexByAddress(bech32.Encode("ln", firstHop[:]))

			var data [32]byte
			outMsg := lnutil.NewMultihopPaymentSetupMsg(firstHopIdx, mh.Amt, msg.HHash, mh.Path, data)
			nd.OmniOut <- outMsg
		}
	}
	return nil
}

func (nd *LitNode) MultihopPaymentSetupHandler(msg lnutil.MultihopPaymentSetupMsg) error {
	return nil
}

func (nd *LitNode) MultihopPaymentSettleHandler(msg lnutil.MultihopPaymentSettleMsg) error {
	return nil
}
