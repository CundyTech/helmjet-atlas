package main

import (
	"fmt"
	"helmjet-atlas/internal/config"
	"log"
	"os"

	"helmjet-atlas/internal/server"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load environment variables
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "helmjet-atlas"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Build integrations config from environment
	integrationsCfg := config.IntegrationsConfig{
		NATSURL:     os.Getenv("NATS_URL"),
		NATSCluster: os.Getenv("NATS_CLUSTER"),
		Kubeconfig:  os.Getenv("KUBECONFIG_PATH"),
	}

	// Connect to MongoDB
	db, err := config.NewMongoDB(config.MongoConfig{
		URI:      mongoURI,
		Database: mongoDB,
	})
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Create indexes
	if err := config.CreateIndexes(db); err != nil {
		log.Fatalf("Failed to create indexes: %v", err)
	}

	// Setup Gin router
	router := gin.Default()
	// Add CORS middleware
	router.Use(corsMiddleware())
	// Setup API routes
	server.SetupRoutes(router, db, integrationsCfg)

	// Start server
	fmt.Printf("Starting server on port %s\n", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
