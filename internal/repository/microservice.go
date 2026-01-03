package repository

import (
	"context"
	"helmjet-atlas/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MicroserviceRepository handles Microservice CRUD operations
type MicroserviceRepository struct {
	collection *mongo.Collection
}

func NewMicroserviceRepository(db *mongo.Database) *MicroserviceRepository {
	return &MicroserviceRepository{collection: db.Collection("microservices")}
}

func (r *MicroserviceRepository) Create(ctx context.Context, service *models.Microservice) (primitive.ObjectID, error) {
	service.ID = primitive.NewObjectID()
	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()
	result, err := r.collection.InsertOne(ctx, service)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return result.InsertedID.(primitive.ObjectID), nil
}

func (r *MicroserviceRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Microservice, error) {
	var service models.Microservice
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&service)
	return &service, err
}

func (r *MicroserviceRepository) List(ctx context.Context, filter bson.M, limit, skip int64) ([]models.Microservice, error) {
	opts := options.Find().SetLimit(limit).SetSkip(skip)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	var services []models.Microservice
	if err = cursor.All(ctx, &services); err != nil {
		return nil, err
	}
	return services, nil
}

func (r *MicroserviceRepository) Update(ctx context.Context, id primitive.ObjectID, service *models.Microservice) error {
	service.UpdatedAt = time.Now()
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": service})
	return err
}

func (r *MicroserviceRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// UpsertByNameNamespace updates existing service by name+namespace or inserts new one
func (r *MicroserviceRepository) UpsertByNameNamespace(ctx context.Context, service *models.Microservice) error {
	now := time.Now()
	service.LastSynced = now
	var existing models.Microservice
	err := r.collection.FindOne(ctx, bson.M{"name": service.Name, "namespace": service.Namespace}).Decode(&existing)
	if err == nil {
		service.ID = existing.ID
		service.UpdatedAt = now
		_, err = r.collection.UpdateOne(ctx, bson.M{"_id": existing.ID}, bson.M{"$set": service})
		return err
	}
	_, err = r.Create(ctx, service)
	return err
}
