package pbacme

import (
	"time"

	firecore "github.com/streamingfast/firehose-core"
)

var _ firecore.Block = (*Block)(nil)

func (b *Block) GetFirehoseBlockID() string {
	return b.Header.Hash
}

func (b *Block) GetFirehoseBlockNumber() uint64 {
	return b.Header.Height
}

func (b *Block) GetFirehoseBlockParentID() string {
	if b.Header.PreviousHash == nil {
		return ""
	}

	return *b.Header.PreviousHash
}

func (b *Block) GetFirehoseBlockParentNumber() uint64 {
	if b.Header.PreviousNum == nil {
		return 0
	}

	return *b.Header.PreviousNum
}

func (b *Block) GetFirehoseBlockTime() time.Time {
	return time.Unix(0, int64(b.Header.Timestamp)).UTC()
}

func (b *Block) GetFirehoseBlockLIBNum() uint64 {
	return b.Header.FinalNum
}
