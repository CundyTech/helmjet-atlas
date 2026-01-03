package repository

import (
	"context"
	"helmjet-atlas/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MicroserviceStore defines storage operations used for microservices
type MicroserviceStore interface {
	Create(ctx context.Context, service *models.Microservice) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Microservice, error)
	List(ctx context.Context, filter bson.M, limit, skip int64) ([]models.Microservice, error)
	Update(ctx context.Context, id primitive.ObjectID, service *models.Microservice) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	UpsertByNameNamespace(ctx context.Context, service *models.Microservice) error
}

// NATSStreamStore defines storage operations used for NATS streams
type NATSStreamStore interface {
	Create(ctx context.Context, stream *models.NATSStream) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.NATSStream, error)
	List(ctx context.Context) ([]models.NATSStream, error)
	Update(ctx context.Context, id primitive.ObjectID, stream *models.NATSStream) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	UpsertByNameCluster(ctx context.Context, stream *models.NATSStream) error
}

// NATSConsumerStore defines storage operations used for NATS consumers
type NATSConsumerStore interface {
	Create(ctx context.Context, consumer *models.NATSConsumer) (primitive.ObjectID, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.NATSConsumer, error)
	List(ctx context.Context) ([]models.NATSConsumer, error)
	Update(ctx context.Context, id primitive.ObjectID, consumer *models.NATSConsumer) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	UpsertByNameStreamCluster(ctx context.Context, consumer *models.NATSConsumer) error
}
