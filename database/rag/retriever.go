package rag

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/tahminator/go-react-template/database"
	"google.golang.org/genai"
)

type SimilarChunk struct {
	Source          string  `db:"source"`
	Chunk           string  `db:"chunk"`
	Distance        float64 `db:"distance"`
	FileType        string  `db:"file_type"`
	ConflictSection string  `db:"conflict_section"`
	LineStart       int     `db:"line_start"`
	LineEnd         int     `db:"line_end"`
	ChunkType       string  `db:"chunk_type"`
}

func getQueryEmbedding(ctx context.Context, query string) ([]float64, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	content := genai.NewContentFromText(query, genai.RoleUser)
	embedding, err := client.Models.EmbedContent(ctx, EmbedModel, []*genai.Content{content}, &genai.EmbedContentConfig{
		TaskType: "RETRIEVAL_QUERY",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	embeddingValues := make([]float64, len(embedding.Embeddings[0].Values))
	for i, val := range embedding.Embeddings[0].Values {
		embeddingValues[i] = float64(val)
	}

	return embeddingValues, nil
}

func SimilarChunks(ctx context.Context, query, repoHash string, k int) ([]SimilarChunk, error) {
	queryEmbedding, err := getQueryEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Convert embedding to pgvector format
	embeddingValues := make([]string, len(queryEmbedding))
	for i, val := range queryEmbedding {
		embeddingValues[i] = fmt.Sprintf("%.6f", val)
	}
	embeddingStr := "[" + strings.Join(embeddingValues, ",") + "]"

	pool, err := database.GetPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get database pool: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT source, chunk, embedding <-> $1 AS distance, file_type, conflict_section, line_start, line_end, chunk_type
		FROM repo_chunks
		WHERE repo_hash = $2
		ORDER BY embedding <-> $1
		LIMIT $3
	`, embeddingStr, repoHash, k)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar chunks: %w", err)
	}
	defer rows.Close()

	var chunks []SimilarChunk
	for rows.Next() {
		var chunk SimilarChunk
		err := rows.Scan(&chunk.Source, &chunk.Chunk, &chunk.Distance, &chunk.FileType, &chunk.ConflictSection, &chunk.LineStart, &chunk.LineEnd, &chunk.ChunkType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk row: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return chunks, nil
}

func SimilarChunksWithThreshold(ctx context.Context, query, repoHash string, k int, maxDistance float64) ([]SimilarChunk, error) {
	queryEmbedding, err := getQueryEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Convert embedding to pgvector format
	embeddingValues := make([]string, len(queryEmbedding))
	for i, val := range queryEmbedding {
		embeddingValues[i] = fmt.Sprintf("%.6f", val)
	}
	embeddingStr := "[" + strings.Join(embeddingValues, ",") + "]"

	pool, err := database.GetPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get database pool: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT source, chunk, embedding <-> $1 AS distance, file_type, conflict_section, line_start, line_end, chunk_type
		FROM repo_chunks
		WHERE repo_hash = $2 AND embedding <-> $1 < $4
		ORDER BY embedding <-> $1
		LIMIT $3
	`, embeddingStr, repoHash, k, maxDistance)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar chunks: %w", err)
	}
	defer rows.Close()

	var chunks []SimilarChunk
	for rows.Next() {
		var chunk SimilarChunk
		err := rows.Scan(&chunk.Source, &chunk.Chunk, &chunk.Distance, &chunk.FileType, &chunk.ConflictSection, &chunk.LineStart, &chunk.LineEnd, &chunk.ChunkType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan chunk row: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return chunks, nil
}

func GetRepoChunksCount(ctx context.Context, repoHash string) (int, error) {
	pool, err := database.GetPool()
	if err != nil {
		return 0, fmt.Errorf("failed to get database pool: %w", err)
	}

	var count int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM repo_chunks WHERE repo_hash = $1", repoHash).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count repo chunks: %w", err)
	}

	return count, nil
}

func DeleteRepoChunks(ctx context.Context, repoHash string) error {
	pool, err := database.GetPool()
	if err != nil {
		return fmt.Errorf("failed to get database pool: %w", err)
	}

	_, err = pool.Exec(ctx, "DELETE FROM repo_chunks WHERE repo_hash = $1", repoHash)
	if err != nil {
		return fmt.Errorf("failed to delete repo chunks: %w", err)
	}

	return nil
}

// GetConflictChunks retrieves chunks specifically related to merge conflicts
func GetConflictChunks(ctx context.Context, repoHash string) ([]SimilarChunk, error) {
	pool, err := database.GetPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get database pool: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT source, chunk, 0.0 AS distance, file_type, conflict_section, line_start, line_end, chunk_type
		FROM repo_chunks
		WHERE repo_hash = $1 AND file_type = 'conflict'
		ORDER BY line_start
	`, repoHash)
	if err != nil {
		return nil, fmt.Errorf("failed to query conflict chunks: %w", err)
	}
	defer rows.Close()

	var chunks []SimilarChunk
	for rows.Next() {
		var chunk SimilarChunk
		err := rows.Scan(&chunk.Source, &chunk.Chunk, &chunk.Distance, &chunk.FileType, &chunk.ConflictSection, &chunk.LineStart, &chunk.LineEnd, &chunk.ChunkType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan conflict chunk row: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over conflict chunk rows: %w", err)
	}

	return chunks, nil
}

// GetContextChunks retrieves context chunks around conflicts for better understanding
func GetContextChunks(ctx context.Context, repoHash, source string, lineStart, lineEnd int, contextLines int) ([]SimilarChunk, error) {
	pool, err := database.GetPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get database pool: %w", err)
	}

	startBound := lineStart - contextLines
	endBound := lineEnd + contextLines

	rows, err := pool.Query(ctx, `
		SELECT source, chunk, 0.0 AS distance, file_type, conflict_section, line_start, line_end, chunk_type
		FROM repo_chunks
		WHERE repo_hash = $1 AND source = $2 
		AND (
			(line_start >= $3 AND line_start <= $4) OR
			file_type = 'context'
		)
		ORDER BY line_start
	`, repoHash, source, startBound, endBound)
	if err != nil {
		return nil, fmt.Errorf("failed to query context chunks: %w", err)
	}
	defer rows.Close()

	var chunks []SimilarChunk
	for rows.Next() {
		var chunk SimilarChunk
		err := rows.Scan(&chunk.Source, &chunk.Chunk, &chunk.Distance, &chunk.FileType, &chunk.ConflictSection, &chunk.LineStart, &chunk.LineEnd, &chunk.ChunkType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan context chunk row: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over context chunk rows: %w", err)
	}

	return chunks, nil
}

// GetSimilarFunctions retrieves similar function definitions for conflict resolution
func GetSimilarFunctions(ctx context.Context, query, repoHash string, k int) ([]SimilarChunk, error) {
	queryEmbedding, err := getQueryEmbedding(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get query embedding: %w", err)
	}

	// Convert embedding to pgvector format
	embeddingValues := make([]string, len(queryEmbedding))
	for i, val := range queryEmbedding {
		embeddingValues[i] = fmt.Sprintf("%.6f", val)
	}
	embeddingStr := "[" + strings.Join(embeddingValues, ",") + "]"

	pool, err := database.GetPool()
	if err != nil {
		return nil, fmt.Errorf("failed to get database pool: %w", err)
	}

	rows, err := pool.Query(ctx, `
		SELECT source, chunk, embedding <-> $1 AS distance, file_type, conflict_section, line_start, line_end, chunk_type
		FROM repo_chunks
		WHERE repo_hash = $2 AND chunk_type = 'function'
		ORDER BY embedding <-> $1
		LIMIT $3
	`, embeddingStr, repoHash, k)
	if err != nil {
		return nil, fmt.Errorf("failed to query similar functions: %w", err)
	}
	defer rows.Close()

	var chunks []SimilarChunk
	for rows.Next() {
		var chunk SimilarChunk
		err := rows.Scan(&chunk.Source, &chunk.Chunk, &chunk.Distance, &chunk.FileType, &chunk.ConflictSection, &chunk.LineStart, &chunk.LineEnd, &chunk.ChunkType)
		if err != nil {
			return nil, fmt.Errorf("failed to scan function chunk row: %w", err)
		}
		chunks = append(chunks, chunk)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over function chunk rows: %w", err)
	}

	return chunks, nil
}
