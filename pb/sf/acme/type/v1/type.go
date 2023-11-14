package pbacme

import (
	"time"
)

func (b *Block) GetFirehoseBlockID() string {
	return b.Hash
}

func (b *Block) GetFirehoseBlockNumber() uint64 {
	return b.Height
}

func (b *Block) GetFirehoseBlockParentNumber() uint64 {
	// This needs to be adapted for your own chain rules!
	return b.Height - 1
}

func (b *Block) GetFirehoseBlockParentID() string {
	return b.PrevHash
}

func (b *Block) GetFirehoseBlockTime() time.Time {
	return time.Unix(0, int64(b.Timestamp)).UTC()
}

func (b *Block) GetFirehoseBlockVersion() int32 {
	// This needs to be adapted for your own version used in pbbstream
	return 0
}

func (b *Block) GetFirehoseBlockLIBNum() uint64 {
	if b.Height == 0 {
		return 0
	}

	// This needs to be adapted for your own chain rules!
	return b.Height - 1
}
