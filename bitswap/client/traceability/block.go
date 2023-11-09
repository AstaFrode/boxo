package traceability

import (
	"time"

	"github.com/AstaFrode/go-libp2p/core/peer"
	blocks "github.com/ipfs/go-block-format"
)

// Block is a block whose provenance has been tracked.
type Block struct {
	blocks.Block

	// From contains the peer id of the node who sent us the block.
	// It will be the zero value if we did not downloaded this block from the
	// network. (such as by getting the block from NotifyNewBlocks).
	From peer.ID
	// Delay contains how long did we had to wait between when we started being
	// intrested and when we actually got the block.
	Delay time.Duration
}
