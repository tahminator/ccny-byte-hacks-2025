package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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

	// Read conflict text from testRepo/main.go
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

	fmt.Println("2. Testing RAG-enhanced conflict resolution...")
	response, err := geminiService.ResolveMergeConflictsWithRAG(ctx, "resolve all merge conflicts in this Go code", repoHash)
	if err != nil {
		log.Fatalf("Failed to resolve conflicts: %v", err)
	}

	fmt.Println("Generated shell commands:")
	fmt.Println(response)
	fmt.Println()

	fmt.Println("3. Testing semantic search...")
	semanticResponse, err := geminiService.ResolveConflictsWithSemanticSearch(ctx, "fix user validation conflicts", repoHash, 5)
	if err != nil {
		log.Printf("Semantic search failed: %v", err)
	} else {
		fmt.Println("Semantic search result:")
		fmt.Println(semanticResponse)
	}

	fmt.Println("=== Test Complete ===")
}
