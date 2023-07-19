package main

import (
	"github.com/streamingfast/firehose-acme/codec"
	pbacme "github.com/streamingfast/firehose-acme/pb/sf/acme/type/v1"
	firecore "github.com/streamingfast/firehose-core"
)

func main() {
	firecore.Main(&firecore.Chain[*pbacme.Block]{
		ShortName:            "acme",
		LongName:             "Acme",
		ExecutableName:       "dummy-blockchain",
		FullyQualifiedModule: "github.com/streamingfast/firehose-acme",
		Version:              version,

		Protocol:        "ACM",
		ProtocolVersion: 1,

		FirstStreamableBlock: 1,

		BlockFactory:         func() firecore.Block { return new(pbacme.Block) },
		ConsoleReaderFactory: codec.NewConsoleReader,

		Tools: &firecore.ToolsConfig[*pbacme.Block]{
			BlockPrinter: printBlock,
		},
	})
}

// Version value, injected via go build `ldflags` at build time, **must** not be removed or inlined
var version = "dev"
