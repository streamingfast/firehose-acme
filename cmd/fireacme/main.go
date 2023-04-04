package main

import (
	"github.com/streamingfast/firehose-acme/codec"
	nodemanager "github.com/streamingfast/firehose-acme/nodemanager"
	firecore "github.com/streamingfast/firehose-core"
)

// Version value, injected via go build `ldflags` at build time
var version = "dev"

func main() {
	firecore.Main(&firecore.Chain{
		ShortName:            "acme",
		LongName:             "Acme",
		ExecutableName:       "dummy-blockchain",
		FullyQualifiedModule: "github.com/streamingfast/firehose-acme",

		Version: version,

		FirstStreamableBlock:                   1,
		BlockDifferenceThresholdConsideredNear: 15,

		ConsoleReaderFactory:   codec.NewConsoleReader,
		ChainSuperviserFactory: nodemanager.NewSuperviser,

		Tools: &firecore.ToolsConfig{
			BlockPrinter: printBlock,
		},
	})
}
