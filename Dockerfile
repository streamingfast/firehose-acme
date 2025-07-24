ARG FIRECORE_VERSION=v1.10.1
ARG DUMMY_BLOCKCHAIN_VERSION=v1.6.1

FROM ghcr.io/streamingfast/dummy-blockchain:${DUMMY_BLOCKCHAIN_VERSION} AS chain

FROM ghcr.io/streamingfast/firehose-core:${FIRECORE_VERSION}

COPY --from=chain /app/dummy-blockchain /app/dummy-blockchain
