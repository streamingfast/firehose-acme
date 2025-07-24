### StreamingFast Firehose Acme Types

Protobuf definitions for the Firehose Acme chain and the `dummy-blockchain`, both used to show operators how to run the Firehose stack and integrator how a chain's integration could look like.

Foremost, this provides [sf.acme.types.v1.Block](https://buf.build/streamingfast/firehose-acme/docs/main:sf.acme.type.v1#sf.acme.type.v1.Block) which is used for consuming Firehose & Substreams Acme block model.

This is probably the only model you should really care about, the rest are meant for internal communications for Firehose operators.

Useful links:
- Documentation: [https://substreams.streamingfast.io/](https://substreams.streamingfast.io/) ([Firehose Docs](https://firehose.streamingfast.io/))
- Source: [https://github.com/streamingfast/firehose-acme](https://github.com/streamingfast/firehose-acme)