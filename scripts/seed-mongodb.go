package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Data models
// Data models
type Microservice struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name"`
	Namespace string             `bson:"namespace"`
	Cluster   string             `bson:"cluster"`
	Image     string             `bson:"image"`
	Replicas  int32              `bson:"replicas"`
	Status    string             `bson:"status"`
	Labels    map[string]string  `bson:"labels"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

type StreamState struct {
	Messages  uint64 `bson:"messages"`
	Bytes     uint64 `bson:"bytes"`
	FirstSeq  uint64 `bson:"firstSeq"`
	LastSeq   uint64 `bson:"lastSeq"`
	Consumers int    `bson:"consumers"`
}

type NATSStream struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	ServiceID primitive.ObjectID `bson:"serviceId"` // Parent Service
	Name      string             `bson:"name"`
	Cluster   string             `bson:"cluster"`
	Subjects  []string           `bson:"subjects"`
	Replicas  int32              `bson:"replicas"`
	Storage   string             `bson:"storage"`
	State     *StreamState       `bson:"state,omitempty"`
	CreatedAt time.Time          `bson:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt"`
}

type NATSConsumer struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	ServiceID     primitive.ObjectID `bson:"serviceId"` // Parent Service (Subscriber)
	StreamID      primitive.ObjectID `bson:"streamId"`  // Source Stream
	Name          string             `bson:"name"`
	StreamName    string             `bson:"streamName"`
	Cluster       string             `bson:"cluster"`
	ConsumerGroup string             `bson:"consumerGroup"`
	Status        string             `bson:"status"`
	Subjects      []string           `bson:"subjects"`
	CreatedAt     time.Time          `bson:"createdAt"`
	UpdatedAt     time.Time          `bson:"updatedAt"`
}

func main() {
	// Get environment variables
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "helmjet-atlas"
	}

	// Connect to MongoDB
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	db := client.Database(mongoDB)

	// Clear existing data
	fmt.Println("Clearing existing collections...")
	collections := []string{
		"microservices",
		"nats_streams",
		"nats_consumers",
		"service_stream_links",   // Remove legacy
		"consumer_service_links", // Remove legacy
		"stream_consumer_links",  // Remove legacy
	}

	for _, collName := range collections {
		coll := db.Collection(collName)
		if err := coll.Drop(context.Background()); err != nil {
			log.Printf("Warning: Failed to drop collection %s: %v\n", collName, err)
		}
	}

	// Create mock data
	fmt.Println("Seeding database with simplified data...")

	// Create microservices
	microservices := []Microservice{
		{
			Name: "api-gateway", Namespace: "default", Cluster: "production",
			Image: "myregistry/api-gateway:1.2.0", Replicas: 3, Status: "Running",
			Labels:    map[string]string{"app": "api-gateway", "version": "1.2.0"},
			CreatedAt: time.Now().Add(-48 * time.Hour), UpdatedAt: time.Now(),
		},
		{
			Name: "auth-service", Namespace: "default", Cluster: "production",
			Image: "myregistry/auth-service:2.1.0", Replicas: 2, Status: "Running",
			Labels:    map[string]string{"app": "auth-service", "version": "2.1.0"},
			CreatedAt: time.Now().Add(-72 * time.Hour), UpdatedAt: time.Now(),
		},
		{
			Name: "user-service", Namespace: "default", Cluster: "production",
			Image: "myregistry/user-service:1.5.0", Replicas: 4, Status: "Running",
			Labels:    map[string]string{"app": "user-service", "version": "1.5.0"},
			CreatedAt: time.Now().Add(-96 * time.Hour), UpdatedAt: time.Now(),
		},
		{
			Name: "order-service", Namespace: "default", Cluster: "production",
			Image: "myregistry/order-service:1.3.0", Replicas: 3, Status: "Running",
			Labels:    map[string]string{"app": "order-service", "version": "1.3.0"},
			CreatedAt: time.Now().Add(-120 * time.Hour), UpdatedAt: time.Now(),
		},
		{
			Name: "payment-service", Namespace: "default", Cluster: "production",
			Image: "myregistry/payment-service:1.7.0", Replicas: 2, Status: "Running",
			Labels:    map[string]string{"app": "payment-service", "version": "1.7.0"},
			CreatedAt: time.Now().Add(-144 * time.Hour), UpdatedAt: time.Now(),
		},
		{
			Name: "notification-service", Namespace: "default", Cluster: "production",
			Image: "myregistry/notification-service:1.0.0", Replicas: 2, Status: "Running",
			Labels:    map[string]string{"app": "notification-service", "version": "1.0.0"},
			CreatedAt: time.Now().Add(-168 * time.Hour), UpdatedAt: time.Now(),
		},
	}

	msIDs := []primitive.ObjectID{}
	msColl := db.Collection("microservices")
	for _, ms := range microservices {
		ms.ID = primitive.NewObjectID()
		msIDs = append(msIDs, ms.ID)
		if _, err := msColl.InsertOne(context.Background(), ms); err != nil {
			log.Fatalf("Failed to insert microservice: %v", err)
		}
		fmt.Printf("Created microservice: %s\n", ms.Name)
	}

	// Microservice IDs map
	// 0: api-gateway
	// 1: auth-service
	// 2: user-service
	// 3: order-service
	// 4: payment-service
	// 5: notification-service

	// Create NATS streams (Parented to Services)
	streams := []NATSStream{
		{
			Name: "events", ServiceID: msIDs[0], // api-gateway
			Cluster: "default", Subjects: []string{"events.>"}, Replicas: 3, Storage: "file",
			State: &StreamState{Messages: 1523, Bytes: 245600, FirstSeq: 1, LastSeq: 1523, Consumers: 0},
		},
		{
			Name: "orders", ServiceID: msIDs[3], // order-service
			Cluster: "default", Subjects: []string{"orders.created", "orders.updated", "orders.shipped"}, Replicas: 3, Storage: "file",
			State: &StreamState{Messages: 8742, Bytes: 1398720, FirstSeq: 1, LastSeq: 8742, Consumers: 0},
		},
		{
			Name: "payments", ServiceID: msIDs[4], // payment-service
			Cluster: "default", Subjects: []string{"payments.initiated", "payments.completed", "payments.failed"}, Replicas: 3, Storage: "file",
			State: &StreamState{Messages: 5621, Bytes: 899360, FirstSeq: 1, LastSeq: 5621, Consumers: 0},
		},
		{
			Name: "users", ServiceID: msIDs[2], // user-service
			Cluster: "default", Subjects: []string{"users.registered", "users.updated", "users.deleted"}, Replicas: 3, Storage: "file",
			State: &StreamState{Messages: 3204, Bytes: 512640, FirstSeq: 1, LastSeq: 3204, Consumers: 0},
		},
		{
			Name: "notifications", ServiceID: msIDs[5], // notification-service (Example: self-contained stream)
			Cluster: "default", Subjects: []string{"notifications.email", "notifications.sms", "notifications.push"}, Replicas: 3, Storage: "file",
			State: &StreamState{Messages: 12456, Bytes: 1993000, FirstSeq: 1, LastSeq: 12456, Consumers: 0},
		},
	}

	streamIDs := []primitive.ObjectID{}
	streamColl := db.Collection("nats_streams")
	for _, stream := range streams {
		stream.ID = primitive.NewObjectID()
		stream.CreatedAt = time.Now().Add(-24 * time.Hour)
		stream.UpdatedAt = time.Now()
		streamIDs = append(streamIDs, stream.ID)
		if _, err := streamColl.InsertOne(context.Background(), stream); err != nil {
			log.Fatalf("Failed to insert stream: %v", err)
		}
		fmt.Printf("Created stream: %s\n", stream.Name)
	}

	// Stream IDs map
	// 0: events
	// 1: orders
	// 2: payments
	// 3: users
	// 4: notifications

	// Create NATS consumers (Parented to Subscriber Service, Linked to Source Stream)
	consumers := []NATSConsumer{
		{
			Name:      "audit-consumer",
			ServiceID: msIDs[1], StreamID: streamIDs[0], // auth-service -> events
			StreamName: "events", Cluster: "default", ConsumerGroup: "auth-service",
			Status: "active", Subjects: []string{"events.>"},
		},
		{
			Name:      "order-payment-handler",
			ServiceID: msIDs[4], StreamID: streamIDs[1], // payment-service -> orders
			StreamName: "orders", Cluster: "default", ConsumerGroup: "payment-service",
			Status: "active", Subjects: []string{"orders.created"},
		},
		{
			Name:      "payment-status-updater",
			ServiceID: msIDs[3], StreamID: streamIDs[2], // order-service -> payments
			StreamName: "payments", Cluster: "default", ConsumerGroup: "order-service",
			Status: "active", Subjects: []string{"payments.completed"},
		},
		{
			Name:      "user-welcome-mailer",
			ServiceID: msIDs[5], StreamID: streamIDs[3], // notification-service -> users
			StreamName: "users", Cluster: "default", ConsumerGroup: "notification-service",
			Status: "active", Subjects: []string{"users.registered"},
		},
		{
			Name:      "order-email-sender",
			ServiceID: msIDs[5], StreamID: streamIDs[1], // notification-service -> orders
			StreamName: "orders", Cluster: "default", ConsumerGroup: "notification-service",
			Status: "active", Subjects: []string{"orders.created"},
		},
		{
			Name:      "high-value-payment-alert",
			ServiceID: msIDs[5], StreamID: streamIDs[2], // notification-service -> payments
			StreamName: "payments", Cluster: "default", ConsumerGroup: "notification-service",
			Status: "active", Subjects: []string{"payments.completed"},
		},
	}

	consumerColl := db.Collection("nats_consumers")
	for _, consumer := range consumers {
		consumer.ID = primitive.NewObjectID()
		consumer.CreatedAt = time.Now().Add(-24 * time.Hour)
		consumer.UpdatedAt = time.Now()
		if _, err := consumerColl.InsertOne(context.Background(), consumer); err != nil {
			log.Fatalf("Failed to insert consumer: %v", err)
		}
		fmt.Printf("Created consumer: %s\n", consumer.Name)
	}

	fmt.Println("\n✅ Database seeded successfully (Simplified Model)!")
	fmt.Printf("Created:\n")
	fmt.Printf("  - %d microservices\n", len(microservices))
	fmt.Printf("  - %d NATS streams\n", len(streams))
	fmt.Printf("  - %d consumers\n", len(consumers))
}
