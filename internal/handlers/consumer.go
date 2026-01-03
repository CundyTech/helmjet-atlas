package handlers

import (
	"helmjet-atlas/internal/models"
	"helmjet-atlas/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// NATSConsumerHandler handles NATS consumer API endpoints
type NATSConsumerHandler struct {
	repo repository.NATSConsumerStore
}

// NewNATSConsumerHandler creates a new NATS consumer handler
func NewNATSConsumerHandler(repo repository.NATSConsumerStore) *NATSConsumerHandler {
	return &NATSConsumerHandler{repo: repo}
}

// CreateConsumer creates a new NATS consumer
func (h *NATSConsumerHandler) CreateConsumer(c *gin.Context) {
	var consumer models.NATSConsumer
	if err := c.ShouldBindJSON(&consumer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	id, err := h.repo.Create(c.Request.Context(), &consumer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id.Hex()})
}

// GetConsumer retrieves a consumer by ID
func (h *NATSConsumerHandler) GetConsumer(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer ID"})
		return
	}

	consumer, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Consumer not found"})
		return
	}

	c.JSON(http.StatusOK, consumer)
}

// ListConsumers lists all NATS consumers with optional filtering
func (h *NATSConsumerHandler) ListConsumers(c *gin.Context) {
	consumers, err := h.repo.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, consumers)
}

// UpdateConsumer updates a NATS consumer
func (h *NATSConsumerHandler) UpdateConsumer(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer ID"})
		return
	}

	var consumer models.NATSConsumer
	if err := c.ShouldBindJSON(&consumer); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Update(c.Request.Context(), id, &consumer); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consumer updated"})
}

// DeleteConsumer deletes a NATS consumer
func (h *NATSConsumerHandler) DeleteConsumer(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid consumer ID"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Consumer deleted"})
}
