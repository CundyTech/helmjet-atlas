package repository

import (
	"context"
	"helmjet-atlas/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NATSConsumerRepository handles NATSConsumer CRUD operations
type NATSConsumerRepository struct {
	collection *mongo.Collection
}

func NewNATSConsumerRepository(db *mongo.Database) *NATSConsumerRepository {
	return &NATSConsumerRepository{collection: db.Collection("nats_consumers")}
}

func (r *NATSConsumerRepository) Create(ctx context.Context, consumer *models.NATSConsumer) (primitive.ObjectID, error) {
	consumer.ID = primitive.NewObjectID()
	consumer.CreatedAt = time.Now()
	consumer.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, consumer)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *NATSConsumerRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.NATSConsumer, error) {
	var consumer models.NATSConsumer
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&consumer)
	return &consumer, err
}

func (r *NATSConsumerRepository) List(ctx context.Context) ([]models.NATSConsumer, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var consumers []models.NATSConsumer
	if err = cursor.All(ctx, &consumers); err != nil {
		return nil, err
	}
	return consumers, nil
}

func (r *NATSConsumerRepository) Update(ctx context.Context, id primitive.ObjectID, consumer *models.NATSConsumer) error {
	consumer.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": consumer})
	return err
}

func (r *NATSConsumerRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// UpsertByNameStreamCluster updates existing consumer by name+streamName+cluster or inserts new one
func (r *NATSConsumerRepository) UpsertByNameStreamCluster(ctx context.Context, consumer *models.NATSConsumer) error {
	now := time.Now()
	consumer.LastSynced = now
	var existing models.NATSConsumer
	err := r.collection.FindOne(ctx, bson.M{"name": consumer.Name, "streamName": consumer.StreamName, "cluster": consumer.Cluster}).Decode(&existing)
	if err == nil {
		consumer.ID = existing.ID
		consumer.UpdatedAt = now
		_, err = r.collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": consumer})
		return err
	}
	_, err = r.Create(ctx, consumer)
	return err
}
