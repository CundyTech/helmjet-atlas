package integrations

import "helmjet-atlas/internal/repository"

// Integrations bundles storage backends used by integration syncs
type Integrations struct {
	MS        repository.MicroserviceStore
	Streams   repository.NATSStreamStore
	Consumers repository.NATSConsumerStore
}

// New creates a new Integrations instance with the provided stores
func New(ms repository.MicroserviceStore, streams repository.NATSStreamStore, consumers repository.NATSConsumerStore) *Integrations {
	return &Integrations{MS: ms, Streams: streams, Consumers: consumers}
}
