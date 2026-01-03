package repository

import (
	"context"
	"helmjet-atlas/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// NATSStreamRepository handles NATSStream CRUD operations
type NATSStreamRepository struct {
	collection *mongo.Collection
}

func NewNATSStreamRepository(db *mongo.Database) *NATSStreamRepository {
	return &NATSStreamRepository{collection: db.Collection("nats_streams")}
}

func (r *NATSStreamRepository) Create(ctx context.Context, stream *models.NATSStream) (primitive.ObjectID, error) {
	stream.ID = primitive.NewObjectID()
	stream.CreatedAt = time.Now()
	stream.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, stream)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *NATSStreamRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.NATSStream, error) {
	var stream models.NATSStream
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&stream)
	return &stream, err
}

func (r *NATSStreamRepository) List(ctx context.Context) ([]models.NATSStream, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	var streams []models.NATSStream
	if err = cursor.All(ctx, &streams); err != nil {
		return nil, err
	}
	return streams, nil
}

func (r *NATSStreamRepository) Update(ctx context.Context, id primitive.ObjectID, stream *models.NATSStream) error {
	stream.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": stream})
	return err
}

func (r *NATSStreamRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// UpsertByNameCluster updates existing stream by name+cluster or inserts new one
func (r *NATSStreamRepository) UpsertByNameCluster(ctx context.Context, stream *models.NATSStream) error {
	now := time.Now()
	stream.LastSynced = now
	var existing models.NATSStream
	err := r.collection.FindOne(ctx, bson.M{"name": stream.Name, "cluster": stream.Cluster}).Decode(&existing)
	if err == nil {
		stream.ID = existing.ID
		stream.UpdatedAt = now
		_, err = r.collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": stream})
		return err
	}
	_, err = r.Create(ctx, stream)
	return err
}
