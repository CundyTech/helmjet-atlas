package config

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI      string
	Database string
}

// IntegrationsConfig holds default settings for external integrations
type IntegrationsConfig struct {
	NATSURL     string
	NATSCluster string
	Kubeconfig  string
}

func NewMongoDB(cfg MongoConfig) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to MongoDB")
	return client.Database(cfg.Database), nil
}

func CreateIndexes(db *mongo.Database) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	indexes := map[string][]bson.D{
		"microservices": {
			bson.D{{Key: "cluster", Value: 1}, {Key: "namespace", Value: 1}, {Key: "name", Value: 1}},
			bson.D{{Key: "lastSyncedAt", Value: 1}},
		},
		"nats_streams": {
			bson.D{{Key: "name", Value: 1}, {Key: "cluster", Value: 1}},
			bson.D{{Key: "cluster", Value: 1}},
		},
		"nats_consumers": {
			bson.D{{Key: "streamName", Value: 1}, {Key: "name", Value: 1}, {Key: "cluster", Value: 1}},
			bson.D{{Key: "status", Value: 1}},
		},
		"service_stream_links": {
			bson.D{{Key: "sourceServiceId", Value: 1}},
			bson.D{{Key: "targetStreamId", Value: 1}},
			bson.D{{Key: "linkType", Value: 1}},
		},
		"consumer_service_links": {
			bson.D{{Key: "serviceId", Value: 1}},
			bson.D{{Key: "consumerId", Value: 1}},
		},
	}

	for collName, indexDocs := range indexes {
		coll := db.Collection(collName)
		for _, indexDoc := range indexDocs {
			_, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: indexDoc})
			if err != nil {
				return fmt.Errorf("failed to create index on %s: %w", collName, err)
			}
		}
	}

	fmt.Println("Created all indexes")
	return nil
}
