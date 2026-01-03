package handlers

import (
	"context"
	"encoding/json"
	"helmjet-atlas/internal/config"
	"helmjet-atlas/internal/integrations"
	"helmjet-atlas/internal/models"
	"helmjet-atlas/internal/repository"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

type IntegrationHandler struct {
	integr *integrations.Integrations
	cfg    config.IntegrationsConfig
}

func NewIntegrationHandler(msRepo repository.MicroserviceStore, streamRepo repository.NATSStreamStore, consumerRepo repository.NATSConsumerStore, cfg config.IntegrationsConfig) *IntegrationHandler {
	integr := integrations.New(msRepo, streamRepo, consumerRepo)
	return &IntegrationHandler{integr: integr, cfg: cfg}
}

// SyncNATS triggers a one-shot sync of NATS JetStream metadata into MongoDB
func (h *IntegrationHandler) SyncNATS(c *gin.Context) {
	var req struct {
		NATSURL string `json:"nats_url"`
		Cluster string `json:"cluster"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// allow empty body
	}
	natsURL := req.NATSURL
	if natsURL == "" {
		natsURL = c.GetHeader("X-NATS-URL")
	}
	if natsURL == "" {
		if h.cfg.NATSURL != "" {
			natsURL = h.cfg.NATSURL
		} else {
			natsURL = "nats://nats:4222"
		}
	}

	cluster := req.Cluster
	if cluster == "" {
		if h.cfg.NATSCluster != "" {
			cluster = h.cfg.NATSCluster
		} else {
			cluster = "default"
		}
	}

	if err := h.integr.SyncNATSOnce(context.Background(), natsURL, cluster); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "nats sync triggered"})
}

// SyncK8s triggers a one-shot sync of Kubernetes Services into MongoDB
func (h *IntegrationHandler) SyncK8s(c *gin.Context) {
	var req struct {
		Kubeconfig string   `json:"kubeconfig"`
		Namespaces []string `json:"namespaces"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		// ignore
	}
	kubeconfig := req.Kubeconfig
	if kubeconfig == "" {
		kubeconfig = c.GetHeader("X-KUBECONFIG")
	}
	if kubeconfig == "" && h.cfg.Kubeconfig != "" {
		kubeconfig = h.cfg.Kubeconfig
	}

	namespaces := req.Namespaces

	if err := h.integr.SyncK8sOnce(context.Background(), kubeconfig, namespaces); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "k8s sync triggered"})
}

// UploadK8s accepts a JSON file (multipart/form-data field `file`) or raw JSON body
// containing an array of Microservice objects and upserts them into storage.
func (h *IntegrationHandler) UploadK8s(c *gin.Context) {
	ctx := c.Request.Context()

	var reader io.Reader
	// Try multipart file first
	f, err := c.FormFile("file")
	if err == nil && f != nil {
		rf, err := f.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
			return
		}
		defer rf.Close()
		reader = rf
	} else {
		// Fallback to raw body
		reader = c.Request.Body
	}

	var services []models.Microservice
	dec := json.NewDecoder(reader)
	if err := dec.Decode(&services); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	count := 0
	for _, svc := range services {
		s := svc // copy
		if err := h.integr.MS.UpsertByNameNamespace(ctx, &s); err != nil {
			// continue on error
			continue
		}
		count++
	}

	c.JSON(http.StatusOK, gin.H{"message": "k8s upload processed", "processed": count})
}

// UploadNATS accepts a JSON file (multipart/form-data field `file`) or raw JSON body
// containing either {"streams": [...], "consumers": [...]} and upserts them.
func (h *IntegrationHandler) UploadNATS(c *gin.Context) {
	ctx := c.Request.Context()
	var reader io.Reader
	f, err := c.FormFile("file")
	if err == nil && f != nil {
		rf, err := f.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to open uploaded file"})
			return
		}
		defer rf.Close()
		reader = rf
	} else {
		reader = c.Request.Body
	}

	// Read into a generic map first
	var payload map[string]json.RawMessage
	dec := json.NewDecoder(reader)
	if err := dec.Decode(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	procStreams := 0
	procConsumers := 0

	if v, ok := payload["streams"]; ok {
		var streams []models.NATSStream
		if err := json.Unmarshal(v, &streams); err == nil {
			for _, st := range streams {
				s := st
				_ = h.integr.Streams.UpsertByNameCluster(ctx, &s)
				procStreams++
			}
		}
	}
	if v, ok := payload["consumers"]; ok {
		var consumers []models.NATSConsumer
		if err := json.Unmarshal(v, &consumers); err == nil {
			for _, co := range consumers {
				cobj := co
				_ = h.integr.Consumers.UpsertByNameStreamCluster(ctx, &cobj)
				procConsumers++
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "nats upload processed", "streams": procStreams, "consumers": procConsumers})
}
