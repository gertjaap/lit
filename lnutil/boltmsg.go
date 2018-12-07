package lnutil

import (
	"bufio"
	"bytes"

	"github.com/mit-dci/lit/lnwire"
)

// BoltMsg is a bridge between LitMsg and Bolt-compatible message from lnwire
type BoltMsg struct {
	InnerMsg lnwire.Message
	PeerIdx  uint32
}

func (msg BoltMsg) Peer() uint32 {
	return msg.PeerIdx
}

func (msg BoltMsg) MsgType() uint16 {
	return uint16(msg.InnerMsg.MsgType())
}

func (msg BoltMsg) Bytes() []byte {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	_ = msg.InnerMsg.Encode(writer, 0)
	return b.Bytes()
}
