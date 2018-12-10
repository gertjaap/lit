package lnp2p

import (
	"bufio"
	"bytes"

	"github.com/mit-dci/lit/lnwire"
)

// BoltMsg is a bridge between lnp2p.Message and Bolt-compatible message from lnwire
type BoltMsg struct {
	InnerMsg lnwire.Message
}

func (msg BoltMsg) Type() uint16 {
	return uint16(msg.InnerMsg.MsgType())
}

func (msg BoltMsg) Bytes() []byte {
	var b bytes.Buffer
	writer := bufio.NewWriter(&b)
	_ = msg.InnerMsg.Encode(writer, 0)
	writer.Flush()
	return b.Bytes()
}
