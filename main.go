package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/samrato/magicstream/database"
	"github.com/samrato/magicstream/routes"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	// Initialize Gin router
	router := gin.Default()

	// Simple health check
	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, MagicStreamMovies!")
	})

	// Configure CORS
	allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
	var origins []string
	if allowedOrigins != "" {
		origins = strings.Split(allowedOrigins, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
			log.Println("Allowed Origin:", origins[i])
		}
	} else {
		origins = []string{"http://localhost:5173"}
		log.Println("Allowed Origin: http://localhost:5173")
	}

	corsConfig := cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Connect to MongoDB
	client := database.Connect()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect MongoDB: %v", err)
		}
	}()
	log.Println("MongoDB connected successfully")

	// Setup routes
	routes.MovieRoutes(router, client)
	routes.UserRoutes(router, client)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s...", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
