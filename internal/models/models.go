package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Microservice represents a Kubernetes microservice
type Microservice struct {
	ID         primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name       string             `bson:"name" json:"name"`
	Namespace  string             `bson:"namespace" json:"namespace"`
	Cluster    string             `bson:"cluster" json:"cluster"`
	Image      string             `bson:"image" json:"image"`
	Replicas   int32              `bson:"replicas" json:"replicas"`
	Status     string             `bson:"status" json:"status"` // Running, Pending, Failed, Unknown
	Labels     map[string]string  `bson:"labels" json:"labels"`
	Ports      []Port             `bson:"ports" json:"ports"`
	CreatedAt  time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time          `bson:"updatedAt" json:"updatedAt"`
	LastSynced time.Time          `bson:"lastSyncedAt" json:"lastSyncedAt"`
}

// Port represents a container port
type Port struct {
	Name          string `bson:"name" json:"name"`
	ContainerPort int32  `bson:"containerPort" json:"containerPort"`
	Protocol      string `bson:"protocol" json:"protocol"` // TCP, UDP
}

// NATSStream represents a NATS JetStream stream
type NATSStream struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceID primitive.ObjectID `bson:"serviceId" json:"serviceId"` // Parent Service
	Name      string             `bson:"name" json:"name"`
	Cluster   string             `bson:"cluster" json:"cluster"`
	Subjects  []string           `bson:"subjects" json:"subjects"`
	Retention Retention          `bson:"retention" json:"retention"`
	Replicas  int32              `bson:"replicas" json:"replicas"`
	Storage   string             `bson:"storage" json:"storage"` // file, memory
	State     *StreamState       `bson:"state,omitempty" json:"state,omitempty"`
	// Health fields computed during sync
	HealthScore  int       `bson:"healthScore,omitempty" json:"healthScore,omitempty"`
	HealthStatus string    `bson:"healthStatus,omitempty" json:"healthStatus,omitempty"`
	UsagePct     float64   `bson:"usagePct,omitempty" json:"usagePct,omitempty"`
	Warnings     []string  `bson:"warnings,omitempty" json:"warnings,omitempty"`
	CreatedAt    time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time `bson:"updatedAt" json:"updatedAt"`
	LastSynced   time.Time `bson:"lastSyncedAt" json:"lastSyncedAt"`
}

// StreamState represents the current state of a NATS stream
type StreamState struct {
	Messages  uint64 `bson:"messages" json:"messages"`   // Total message count
	Bytes     uint64 `bson:"bytes" json:"bytes"`         // Total bytes
	FirstSeq  uint64 `bson:"firstSeq" json:"firstSeq"`   // First sequence number
	LastSeq   uint64 `bson:"lastSeq" json:"lastSeq"`     // Last sequence number
	Consumers int    `bson:"consumers" json:"consumers"` // Number of consumers
}

// Retention policy for NATS streams
type Retention struct {
	Policy   string `bson:"policy" json:"policy"` // limits, interest, workqueue
	MaxAge   int64  `bson:"maxAge" json:"maxAge"` // milliseconds
	MaxBytes int64  `bson:"maxBytes" json:"maxBytes"`
	MaxMsgs  int64  `bson:"maxMsgs" json:"maxMsgs"`
}

// NATSConsumer represents a NATS JetStream consumer
type NATSConsumer struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	ServiceID      primitive.ObjectID `bson:"serviceId" json:"serviceId"` // Parent Service (Subscriber)
	StreamID       primitive.ObjectID `bson:"streamId" json:"streamId"`   // Source Stream
	Name           string             `bson:"name" json:"name"`
	StreamName     string             `bson:"streamName" json:"streamName"`
	Cluster        string             `bson:"cluster" json:"cluster"`
	ConsumerGroup  string             `bson:"consumerGroup" json:"consumerGroup"`
	Subjects       []string           `bson:"subjects" json:"subjects"`
	DeliveryPolicy DeliveryPolicy     `bson:"deliveryPolicy" json:"deliveryPolicy"`
	AckPolicy      string             `bson:"ackPolicy" json:"ackPolicy"` // none, all, explicit
	AckWait        int64              `bson:"ackWait" json:"ackWait"`     // milliseconds
	MaxDeliver     int32              `bson:"maxDeliver" json:"maxDeliver"`
	RateLimit      int64              `bson:"rateLimit" json:"rateLimit"` // bytes per second
	Status         string             `bson:"status" json:"status"`       // active, inactive
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
	LastSynced     time.Time          `bson:"lastSyncedAt" json:"lastSyncedAt"`
}

// DeliveryPolicy defines how messages are delivered to consumers
type DeliveryPolicy struct {
	Type      string    `bson:"type" json:"type"` // all, last, new, byStartTime, byStartSeq
	StartTime time.Time `bson:"startTime,omitempty" json:"startTime,omitempty"`
	StartSeq  int64     `bson:"startSeq,omitempty" json:"startSeq,omitempty"`
}

// Metrics represents performance metrics for entities
type Metrics struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	EntityType   string             `bson:"entityType" json:"entityType"` // microservice, stream, consumer
	EntityID     primitive.ObjectID `bson:"entityId" json:"entityId"`
	Cluster      string             `bson:"cluster" json:"cluster"`
	CPUUsage     float64            `bson:"cpuUsage" json:"cpuUsage"`       // percentage
	MemoryUsage  int64              `bson:"memoryUsage" json:"memoryUsage"` // bytes
	ErrorCount   int64              `bson:"errorCount" json:"errorCount"`
	SuccessCount int64              `bson:"successCount" json:"successCount"`
	LastUpdated  time.Time          `bson:"lastUpdated" json:"lastUpdated"`
}

// Cluster represents a Kubernetes or NATS cluster
type Cluster struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string             `bson:"name" json:"name"`
	Type      string             `bson:"type" json:"type"` // kubernetes, nats, hybrid
	Endpoint  string             `bson:"endpoint" json:"endpoint"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
}
