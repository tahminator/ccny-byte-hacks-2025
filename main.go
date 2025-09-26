package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/tahminator/go-react-template/api"
	"github.com/tahminator/go-react-template/database"
	"google.golang.org/genai"
)

const defaultPort = "8080"

//go:embed static/*
var content embed.FS

func main() {
	if os.Getenv("ENV") != "production" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Failed to load .env: %v", err)
		}
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	err := database.Connect()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	db, err := database.GetPool()
	if err != nil {
		log.Fatalf("Failed to get database pool: %v", err)
	}
	defer db.Close()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatalf("GEMINI_API_KEY environment variable is required")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	r := gin.Default()

	api.NewRouter(r, db, client)

	if os.Getenv("ENV") == "production" {
		r.Static("/", "./static")
	}

	// NOTE will bind on all local ip interfaces. pls be careful.
	r.Run(fmt.Sprintf(":%s", port))
}
