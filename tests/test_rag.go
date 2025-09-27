package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/tahminator/go-react-template/api/gemini"
	"github.com/tahminator/go-react-template/database"
	"github.com/tahminator/go-react-template/database/rag"
	"github.com/tahminator/go-react-template/database/repository/repo_chunks"
	"google.golang.org/genai"
)

func main() {
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	testRepoPath := filepath.Join("..", "testRepo", "main.go")
	conflictText, err := ioutil.ReadFile(testRepoPath)
	if err != nil {
		log.Fatalf("Failed to read testRepo/main.go: %v", err)
	}

	fileContent := rag.FileContent{
		Path:    "testRepo/main.go",
		Content: string(conflictText),
	}

	fmt.Println("=== RAG Workflow Test ===")
	fmt.Println()

	fmt.Println("1. Embedding repository...")
	result, err := rag.EmbedRepoPgVector(ctx, []rag.FileContent{fileContent})
	if err != nil {
		log.Fatalf("Failed to embed repository: %v", err)
	}
	fmt.Printf("Result: %s\n", result)

	repoHash := rag.RepoHash([]rag.FileContent{fileContent})
	fmt.Printf("Repository hash: %s\n", repoHash)
	fmt.Println()

	geminiClient, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini client: %v", err)
	}

	pool, err := database.GetPool()
	if err != nil {
		log.Fatalf("Failed to get database pool: %v", err)
	}
	repoChunksRepo := repo_chunks.NewPostgresRepoChunksRepository(pool)
	geminiService := gemini.NewGeminiService(geminiClient, repoChunksRepo)

	fmt.Println("2. Resolving merge conflicts to generate new file...")
	resolvedContent, err := geminiService.ResolveConflictsToFile(ctx, string(conflictText), "testRepo/main.go", "resolve all merge conflicts in this Go code", repoHash)
	if err != nil {
		log.Fatalf("Failed to resolve conflicts: %v", err)
	}

	fmt.Println("Generated resolved file content:")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println(resolvedContent)
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	resolvedFilePath := "../testRepo/main_resolved.go"
	if err := ioutil.WriteFile(resolvedFilePath, []byte(resolvedContent), 0644); err != nil {
		log.Printf("Warning: Failed to write resolved file: %v", err)
	} else {
		fmt.Printf("Resolved file written to: %s\n", resolvedFilePath)
	}
	fmt.Println()

	fmt.Println("3. Testing semantic search resolution...")
	semanticResolved, err := geminiService.ResolveConflictsToFile(ctx, string(conflictText), "testRepo/main.go", "fix user validation conflicts", repoHash)
	if err != nil {
		log.Printf("Semantic search resolution failed: %v", err)
	} else {
		fmt.Println("Semantic search resolved content:")
		fmt.Println("=" + strings.Repeat("=", 50))
		fmt.Println(semanticResolved)
		fmt.Println("=" + strings.Repeat("=", 50))

		semanticFilePath := "../testRepo/main_semantic_resolved.go"
		if err := ioutil.WriteFile(semanticFilePath, []byte(semanticResolved), 0644); err != nil {
			log.Printf("Warning: Failed to write semantic resolved file: %v", err)
		} else {
			fmt.Printf("Semantic resolved file written to: %s\n", semanticFilePath)
		}
	}

	fmt.Println("=== Test Complete ===")
}
