package bitswap_message_pb_test

import (
	"bytes"
	"testing"

	u "github.com/AstaFrode/boxo/util"
	"github.com/ipfs/go-cid"

	pb "github.com/AstaFrode/boxo/bitswap/message/pb"
)

func TestCID(t *testing.T) {
	expected := [...]byte{
		10, 34, 18, 32, 195, 171,
		143, 241, 55, 32, 232, 173,
		144, 71, 221, 57, 70, 107,
		60, 137, 116, 229, 146, 194,
		250, 56, 61, 74, 57, 96,
		113, 76, 174, 240, 196, 242,
	}

	c := cid.NewCidV0(u.Hash([]byte("foobar")))
	msg := pb.Message_BlockPresence{Cid: pb.Cid{Cid: c}}
	actual, err := msg.Marshal()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(actual, expected[:]) {
		t.Fatal("failed to correctly encode custom CID type")
	}
}
