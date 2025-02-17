package internal

import (
	"github.com/AstaFrode/go-libp2p/core/protocol"
)

var (
	// ProtocolBitswapNoVers is equivalent to the legacy bitswap protocol
	ProtocolBitswapNoVers protocol.ID = "/ipfs/bitswap"
	// ProtocolBitswapOneZero is the prefix for the legacy bitswap protocol
	ProtocolBitswapOneZero protocol.ID = "/ipfs/bitswap/1.0.0"
	// ProtocolBitswapOneOne is the the prefix for version 1.1.0
	ProtocolBitswapOneOne protocol.ID = "/ipfs/bitswap/1.1.0"
	// ProtocolBitswap is the current version of the bitswap protocol: 1.2.0
	ProtocolBitswap protocol.ID = "/ipfs/bitswap/1.2.0"
)

var DefaultProtocols = []protocol.ID{
	ProtocolBitswap,
	ProtocolBitswapOneOne,
	ProtocolBitswapOneZero,
	ProtocolBitswapNoVers,
}
