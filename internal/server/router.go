package server

import (
	"helmjet-atlas/internal/config"
	"helmjet-atlas/internal/handlers"
	"helmjet-atlas/internal/repository"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRoutes configures all API routes
func SetupRoutes(router *gin.Engine, db *mongo.Database, integrationsCfg config.IntegrationsConfig) {
	// Initialize repositories
	microserviceRepo := repository.NewMicroserviceRepository(db)
	streamRepo := repository.NewNATSStreamRepository(db)
	consumerRepo := repository.NewNATSConsumerRepository(db)

	// Initialize handlers
	msHandler := handlers.NewMicroserviceHandler(microserviceRepo)
	streamHandler := handlers.NewNATSStreamHandler(streamRepo)
	consumerHandler := handlers.NewNATSConsumerHandler(consumerRepo)
	integrationHandler := handlers.NewIntegrationHandler(microserviceRepo, streamRepo, consumerRepo, integrationsCfg)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Microservices endpoints
		msGroup := v1.Group("/microservices")
		{
			msGroup.POST("", msHandler.CreateMicroservice)
			msGroup.GET("", msHandler.ListMicroservices)
			msGroup.GET("/:id", msHandler.GetMicroservice)
			msGroup.PUT("/:id", msHandler.UpdateMicroservice)
			msGroup.DELETE("/:id", msHandler.DeleteMicroservice)
		}

		// NATS Streams endpoints
		streamsGroup := v1.Group("/streams")
		{
			streamsGroup.POST("", streamHandler.CreateStream)
			streamsGroup.GET("", streamHandler.ListStreams)
			streamsGroup.GET("/:id", streamHandler.GetStream)
			streamsGroup.PUT("/:id", streamHandler.UpdateStream)
			streamsGroup.DELETE("/:id", streamHandler.DeleteStream)
		}

		// NATS Consumers endpoints
		consumersGroup := v1.Group("/consumers")
		{
			consumersGroup.POST("", consumerHandler.CreateConsumer)
			consumersGroup.GET("", consumerHandler.ListConsumers)
			consumersGroup.GET("/:id", consumerHandler.GetConsumer)
			consumersGroup.PUT("/:id", consumerHandler.UpdateConsumer)
			consumersGroup.DELETE("/:id", consumerHandler.DeleteConsumer)
		}

		// Integrations endpoints
		integrationsGroup := v1.Group("/integrations")
		{
			integrationsGroup.POST("/nats/sync", integrationHandler.SyncNATS)
			integrationsGroup.POST("/nats/upload", integrationHandler.UploadNATS)
			integrationsGroup.POST("/k8s/sync", integrationHandler.SyncK8s)
			integrationsGroup.POST("/k8s/upload", integrationHandler.UploadK8s)
		}
	}

	// Simple static serving: serve index at `/` and assets under `/static`.
	staticDir := "./visualization"
	router.GET("/", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "index.html"))
	})
	// Serve the main JS file referenced by index.html
	router.GET("/topology.js", func(c *gin.Context) {
		c.File(filepath.Join(staticDir, "topology.js"))
	})
	// Serve assets (js/css) under /static/*filepath to avoid catch-all conflicts with /api.
	router.Static("/static", staticDir)

	// Serve example pages (runnable demos) under /examples/*filepath
	router.Static("/examples", filepath.Join(staticDir, "examples"))

	// Alias for older path `/example/*` used by some local tests/links
	router.Static("/example", filepath.Join(staticDir, "examples"))
}
