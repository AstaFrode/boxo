package bitswap

import (
	"github.com/AstaFrode/boxo/bitswap/message"
	"github.com/AstaFrode/boxo/bitswap/tracer"
	"github.com/AstaFrode/go-libp2p/core/peer"
)

type sendOnlyTracer interface {
	MessageSent(peer.ID, message.BitSwapMessage)
}

var _ tracer.Tracer = nopReceiveTracer{}

// we need to only trace sends because we already trace receives in the polyfill object (to not get them traced twice)
type nopReceiveTracer struct {
	sendOnlyTracer
}

func (nopReceiveTracer) MessageReceived(peer.ID, message.BitSwapMessage) {}
