# NATS Integration

This document describes the HTTP API and model mappings used by the NATS (JetStream) integration in Helmjet Atlas.

## Overview

The NATS integration queries a NATS JetStream server to discover streams and consumers and persists them into the application's storage using the `NATSStream` and `NATSConsumer` models.

Key interactions with NATS JetStream:

- List streams: `jsm.StreamNames()` / `jsm.StreamInfo(name)`
- List consumers per stream: `jsm.Consumers(streamName)`

The server exposes a simple HTTP endpoint to trigger a one-shot sync.

## API Endpoint

- POST `/api/v1/integrations/nats/sync`

Request body (JSON, optional):

```json
{
  "nats_url": "nats://nats:4222",  // optional, default from env or header
  "cluster": "default"            // optional label used to tag streams/consumers
}
```

Alternatively the request may set the `X-NATS-URL` header to pass the NATS connection string.

Response (200):

```json
{
  "message": "nats sync triggered"
}
```

On error, a 500 is returned with the error in the body.

## Model mappings

Data discovered from JetStream is mapped into the following Go models (see `internal/models/models.go`). Below are the important fields and how they are populated.

### NATSStream

- `name` (string): `info.Config.Name`
- `cluster` (string): the `cluster` value supplied to the sync
- `subjects` ([]string): `info.Config.Subjects`
- `replicas` (int32): `info.Config.Replicas`
- `storage` (string): `info.Config.Storage.String()` (e.g., `file`, `memory`)
- `state` (*StreamState): populated from `info.State` with fields:
  - `messages`: `info.State.Msgs`
  - `bytes`: `info.State.Bytes`
  - `firstSeq`: `info.State.FirstSeq`
  - `lastSeq`: `info.State.LastSeq`
  - `consumers`: currently set to 0 in the sync (consumer counts may be derived separately)
- `retention` (Retention): mapped from `info.Config.Retention`, `MaxAge`, `MaxBytes`, `MaxMsgs`.

Example JSON produced when stored:

```json
{
  "name": "ORDERS",
  "cluster": "default",
  "subjects": ["orders.*"],
  "replicas": 1,
  "storage": "file",
  "state": { "messages": 12345, "bytes": 987654, "firstSeq": 1, "lastSeq": 12345 },
  "retention": { "policy": "limits", "maxAge": 3600000 }
}
```

### NATSConsumer

- `name` (string): consumer name discovered (e.g., `ci.Name`)
- `streamName` (string): parent stream's name
- `cluster` (string): same cluster tag as used for the stream
- `status` (string): textual status (e.g., `active`)
- `consumerGroup`, `deliveryPolicy`, `ackPolicy`, `ackWait`, `maxDeliver`, `rateLimit` — when available these are mapped from the consumer configuration.

Example JSON produced when stored:

```json
{
  "name": "orders-processor",
  "streamName": "ORDERS",
  "cluster": "default",
  "consumerGroup": "processors",
  "subjects": ["orders.*"],
  "status": "active"
}
```

## Notes

- The sync operation is one-shot: it connects to NATS, enumerates streams and consumers, and upserts records using repository helpers.
- Upserts use unique keys such as `(name, cluster)` for streams and `(name, streamName, cluster)` for consumers.
- Authentication, TLS, and advanced NATS configuration should be provided via the `nats_url` (which can include creds or tls schemes) or via environment configuration.

## Vendor documentation

- NATS official docs: https://docs.nats.io/
- JetStream guide: https://docs.nats.io/jetstream
- NATS Go client (`nats.go`) pkg docs: https://pkg.go.dev/github.com/nats-io/nats.go
- NATS JetStream developer guides: https://docs.nats.io/using-nats/developer/jetstream

Refer to these vendor sources when configuring authentication, TLS, or advanced JetStream features used by the sync logic.
