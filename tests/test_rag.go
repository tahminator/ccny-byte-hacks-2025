package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/tahminator/go-react-template/database"
	"github.com/tahminator/go-react-template/database/rag"
)

func main() {
	// Load environment variables from root directory
	if err := godotenv.Load("../.env"); err != nil {
		log.Printf("Warning: Failed to load .env file: %v", err)
	}

	// Check for required environment variables
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY environment variable is required")
	}

	// Connect to database
	if err := database.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	ctx := context.Background()

	// Sample text with merge conflicts (your provided example)
	sampleText := `
	
package main

import "fmt"

<<<<<<< HEAD
func greet() {
    fmt.Println("Hello from main branch")
}
=======
func greet() {
    fmt.Println("Hello from feature branch")
}
>>>>>>> feature/greeting-change

<<<<<<< HEAD
func main() {
    greet()
}
=======
func main() {
    fmt.Println("Starting program...")
    greet()
}
>>>>>>> feature/main-change

`

	// Create FileContent struct for the sample
	fileContent := rag.FileContent{
		Path:    "test/main.go",
		Content: sampleText,
	}

	fmt.Println("=== RAG Embedder and Retriever Test ===")
	fmt.Println()

	// Test 1: Embed the repository
	fmt.Println("1. Testing Embedding...")
	result, err := rag.EmbedRepoPgVector(ctx, []rag.FileContent{fileContent})
	if err != nil {
		log.Fatalf("Failed to embed repository: %v", err)
	}
	fmt.Printf("Embedding result: %s\n", result)
	fmt.Println()

	// Get the repository hash for subsequent queries
	repoHash := rag.RepoHash([]rag.FileContent{fileContent})
	fmt.Printf("Repository hash: %s\n", repoHash)
	fmt.Println()

	// Test 2: Get repository chunks count
	fmt.Println("2. Testing Repository Chunks Count...")
	count, err := rag.GetRepoChunksCount(ctx, repoHash)
	if err != nil {
		log.Fatalf("Failed to get chunks count: %v", err)
	}
	fmt.Printf("Total chunks in repository: %d\n", count)
	fmt.Println()

	// Test 3: Test similarity search with various queries
	fmt.Println("3. Testing Similarity Search...")

	queries := []string{
		"greet function",
		"main function",
		"Hello from main branch",
		"Hello from feature branch",
		"Starting program",
		"merge conflict",
		"import fmt",
		"package main",
	}

	for i, query := range queries {
		fmt.Printf("Query %d: \"%s\"\n", i+1, query)
		chunks, err := rag.SimilarChunks(ctx, query, repoHash, 3)
		if err != nil {
			log.Printf("Failed to get similar chunks for query '%s': %v", query, err)
			continue
		}

		fmt.Printf("Found %d similar chunks:\n", len(chunks))
		for j, chunk := range chunks {
			fmt.Printf("  Chunk %d (distance: %.4f):\n", j+1, chunk.Distance)
			fmt.Printf("    Source: %s\n", chunk.Source)
			fmt.Printf("    File Type: %s\n", chunk.FileType)
			fmt.Printf("    Conflict Section: %s\n", chunk.ConflictSection)
			fmt.Printf("    Chunk Type: %s\n", chunk.ChunkType)
			fmt.Printf("    Lines: %d-%d\n", chunk.LineStart, chunk.LineEnd)
			fmt.Printf("    Content: %.100s...\n", chunk.Chunk)
			fmt.Println()
		}
		fmt.Println("---")
	}

	// Test 4: Test conflict-specific retrieval
	fmt.Println("4. Testing Conflict-Specific Retrieval...")
	conflictChunks, err := rag.GetConflictChunks(ctx, repoHash)
	if err != nil {
		log.Printf("Failed to get conflict chunks: %v", err)
	} else {
		fmt.Printf("Found %d conflict chunks:\n", len(conflictChunks))
		for i, chunk := range conflictChunks {
			fmt.Printf("  Conflict Chunk %d:\n", i+1)
			fmt.Printf("    Source: %s\n", chunk.Source)
			fmt.Printf("    Conflict Section: %s\n", chunk.ConflictSection)
			fmt.Printf("    Lines: %d-%d\n", chunk.LineStart, chunk.LineEnd)
			fmt.Printf("    Content: %.200s...\n", chunk.Chunk)
			fmt.Println()
		}
	}
	fmt.Println()

	// Test 5: Test function-specific retrieval
	fmt.Println("5. Testing Function-Specific Retrieval...")
	funcChunks, err := rag.GetSimilarFunctions(ctx, "greet function", repoHash, 5)
	if err != nil {
		log.Printf("Failed to get similar functions: %v", err)
	} else {
		fmt.Printf("Found %d function chunks:\n", len(funcChunks))
		for i, chunk := range funcChunks {
			fmt.Printf("  Function Chunk %d (distance: %.4f):\n", i+1, chunk.Distance)
			fmt.Printf("    Source: %s\n", chunk.Source)
			fmt.Printf("    Conflict Section: %s\n", chunk.ConflictSection)
			fmt.Printf("    Lines: %d-%d\n", chunk.LineStart, chunk.LineEnd)
			fmt.Printf("    Content: %.200s...\n", chunk.Chunk)
			fmt.Println()
		}
	}
	fmt.Println()

	// Test 6: Test similarity search with threshold
	fmt.Println("6. Testing Similarity Search with Threshold...")
	thresholdChunks, err := rag.SimilarChunksWithThreshold(ctx, "Hello from main branch", repoHash, 5, 0.5)
	if err != nil {
		log.Printf("Failed to get similar chunks with threshold: %v", err)
	} else {
		fmt.Printf("Found %d chunks within threshold (distance < 0.5):\n", len(thresholdChunks))
		for i, chunk := range thresholdChunks {
			fmt.Printf("  Chunk %d (distance: %.4f):\n", i+1, chunk.Distance)
			fmt.Printf("    Source: %s\n", chunk.Source)
			fmt.Printf("    Conflict Section: %s\n", chunk.ConflictSection)
			fmt.Printf("    Content: %.200s...\n", chunk.Chunk)
			fmt.Println()
		}
	}
	fmt.Println()

	// Test 7: Test context retrieval around conflicts
	fmt.Println("7. Testing Context Retrieval Around Conflicts...")
	if len(conflictChunks) > 0 {
		firstConflict := conflictChunks[0]
		contextChunks, err := rag.GetContextChunks(ctx, repoHash, firstConflict.Source, firstConflict.LineStart, firstConflict.LineEnd, 5)
		if err != nil {
			log.Printf("Failed to get context chunks: %v", err)
		} else {
			fmt.Printf("Found %d context chunks around conflict:\n", len(contextChunks))
			for i, chunk := range contextChunks {
				fmt.Printf("  Context Chunk %d:\n", i+1)
				fmt.Printf("    Source: %s\n", chunk.Source)
				fmt.Printf("    File Type: %s\n", chunk.FileType)
				fmt.Printf("    Conflict Section: %s\n", chunk.ConflictSection)
				fmt.Printf("    Lines: %d-%d\n", chunk.LineStart, chunk.LineEnd)
				fmt.Printf("    Content: %.200s...\n", chunk.Chunk)
				fmt.Println()
			}
		}
	}
	fmt.Println()

	// Test 8: Clean up (optional - uncomment if you want to clean up test data)
	// fmt.Println("8. Cleaning up test data...")
	// err = rag.DeleteRepoChunks(ctx, repoHash)
	// if err != nil {
	//     log.Printf("Failed to delete test chunks: %v", err)
	// } else {
	//     fmt.Println("Test data cleaned up successfully")
	// }

	fmt.Println("=== Test Complete ===")
}
