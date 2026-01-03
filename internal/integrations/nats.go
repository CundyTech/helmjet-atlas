package integrations

import (
	"context"
	"fmt"
	"log"

	"helmjet-atlas/internal/models"

	"github.com/nats-io/nats.go"
)

// SyncNATSOnce connects to NATS JetStream and syncs streams and consumers into storage
func (i *Integrations) SyncNATSOnce(ctx context.Context, natsURL, cluster string) error {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer nc.Close()

	jsm, err := nc.JetStream()
	if err != nil {
		return fmt.Errorf("failed to get JetStream context: %w", err)
	}

	if err := i.syncStreams(ctx, jsm, cluster); err != nil {
		log.Printf("syncStreams error: %v", err)
	}
	if err := i.syncConsumers(ctx, jsm, cluster); err != nil {
		log.Printf("syncConsumers error: %v", err)
	}

	return nil
}

func (i *Integrations) syncStreams(ctx context.Context, jsm nats.JetStreamContext, cluster string) error {
	streamsIterator := jsm.StreamNames()
	synced := 0

	for name := range streamsIterator {
		info, err := jsm.StreamInfo(name)
		if err != nil {
			log.Printf("error getting stream info %s: %v", name, err)
			continue
		}

		stream := &models.NATSStream{
			Name:     info.Config.Name,
			Cluster:  cluster,
			Subjects: info.Config.Subjects,
			Replicas: int32(info.Config.Replicas),
			Storage:  info.Config.Storage.String(),
			State: &models.StreamState{
				Messages:  info.State.Msgs,
				Bytes:     info.State.Bytes,
				FirstSeq:  info.State.FirstSeq,
				LastSeq:   info.State.LastSeq,
				Consumers: 0,
			},
		}

		// Retention mapping
		stream.Retention = models.Retention{
			Policy:   fmt.Sprint(info.Config.Retention),
			MaxAge:   info.Config.MaxAge.Nanoseconds() / 1e6,
			MaxBytes: info.Config.MaxBytes,
			MaxMsgs:  info.Config.MaxMsgs,
		}

		// Compute usage and simple health indicators when limits are configured
		var warnings []string
		var usagePct float64
		if info.Config.MaxBytes > 0 {
			usagePct = (float64(info.State.Bytes) / float64(info.Config.MaxBytes)) * 100
			if usagePct >= 95 {
				warnings = append(warnings, "stream storage >= 95% of MaxBytes")
			} else if usagePct >= 80 {
				warnings = append(warnings, "stream storage >= 80% of MaxBytes")
			}
		} else if info.Config.MaxMsgs > 0 {
			usagePct = (float64(info.State.Msgs) / float64(info.Config.MaxMsgs)) * 100
			if usagePct >= 95 {
				warnings = append(warnings, "message count >= 95% of MaxMsgs")
			} else if usagePct >= 80 {
				warnings = append(warnings, "message count >= 80% of MaxMsgs")
			}
		}

		// Determine a basic health score and status
		healthScore := 100
		if usagePct >= 95 {
			healthScore -= 60
		} else if usagePct >= 80 {
			healthScore -= 30
		}

		healthStatus := "Healthy"
		if healthScore < 50 {
			healthStatus = "Critical"
		} else if healthScore < 80 {
			healthStatus = "Warning"
		}

		stream.UsagePct = usagePct
		stream.HealthScore = healthScore
		stream.HealthStatus = healthStatus
		stream.Warnings = warnings

		if err := i.Streams.UpsertByNameCluster(ctx, stream); err != nil {
			log.Printf("failed to upsert stream %s: %v", name, err)
			continue
		}
		synced++
	}

	log.Printf("Synced %d streams", synced)
	return nil
}

func (i *Integrations) syncConsumers(ctx context.Context, jsm nats.JetStreamContext, cluster string) error {
	streamIterator := jsm.StreamNames()
	total := 0

	for streamName := range streamIterator {
		consumersIterator := jsm.Consumers(streamName)
		for ci := range consumersIterator {
			consumer := &models.NATSConsumer{
				Name:       ci.Name,
				StreamName: streamName,
				Cluster:    cluster,
				Status:     "active",
			}

			if err := i.Consumers.UpsertByNameStreamCluster(ctx, consumer); err != nil {
				log.Printf("failed to upsert consumer %s: %v", ci.Name, err)
				continue
			}
			total++
		}
	}

	log.Printf("Synced %d consumers", total)
	return nil
}
