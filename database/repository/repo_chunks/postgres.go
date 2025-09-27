package repo_chunks

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepoChunksRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepoChunksRepository(db *pgxpool.Pool) *PostgresRepoChunksRepository {
	return &PostgresRepoChunksRepository{
		db: db,
	}
}

func (repo *PostgresRepoChunksRepository) InsertChunk(ctx context.Context, chunk *RepoChunk) error {
	embeddingValues := make([]string, len(chunk.Embedding))
	for i, val := range chunk.Embedding {
		embeddingValues[i] = fmt.Sprintf("%.6f", val)
	}
	embeddingStr := "[" + strings.Join(embeddingValues, ",") + "]"

	query := `
		INSERT INTO repo_chunks (repo_hash, source, chunk, embedding, file_type, conflict_section, line_start, line_end, chunk_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := repo.db.Exec(ctx, query,
		chunk.RepoHash, chunk.Source, chunk.Chunk, embeddingStr,
		chunk.FileType, chunk.ConflictSection, chunk.LineStart, chunk.LineEnd, chunk.ChunkType)

	return err
}

func (repo *PostgresRepoChunksRepository) GetConflictChunks(ctx context.Context, repoHash string) ([]SimilarChunk, error) {
	query := `
		SELECT source, chunk, 0.0 AS distance, file_type, conflict_section, line_start, line_end, chunk_type
		FROM repo_chunks
		WHERE repo_hash = $1 AND file_type = 'conflict'
		ORDER BY line_start
	`

	rows, err := repo.db.Query(ctx, query, repoHash)
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

	return chunks, nil
}

func (repo *PostgresRepoChunksRepository) GetContextChunks(ctx context.Context, repoHash, source string, lineStart, lineEnd, contextLines int) ([]SimilarChunk, error) {
	startBound := lineStart - contextLines
	endBound := lineEnd + contextLines

	query := `
		SELECT source, chunk, 0.0 AS distance, file_type, conflict_section, line_start, line_end, chunk_type
		FROM repo_chunks
		WHERE repo_hash = $1 AND source = $2 
		AND (
			(line_start >= $3 AND line_start <= $4) OR
			file_type = 'context'
		)
		ORDER BY line_start
	`

	rows, err := repo.db.Query(ctx, query, repoHash, source, startBound, endBound)
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

	return chunks, nil
}

func (repo *PostgresRepoChunksRepository) GetSimilarChunks(ctx context.Context, query, repoHash string, k int) ([]SimilarChunk, error) {
	// This would need the embedding vector for similarity search
	// For now, return empty - this should be implemented with proper embedding search
	return []SimilarChunk{}, nil
}

func (repo *PostgresRepoChunksRepository) GetSimilarChunksWithThreshold(ctx context.Context, query, repoHash string, k int, maxDistance float64) ([]SimilarChunk, error) {
	// This would need the embedding vector for similarity search
	// For now, return empty - this should be implemented with proper embedding search
	return []SimilarChunk{}, nil
}

func (repo *PostgresRepoChunksRepository) GetSimilarFunctions(ctx context.Context, query, repoHash string, k int) ([]SimilarChunk, error) {
	// This would need the embedding vector for similarity search
	// For now, return empty - this should be implemented with proper embedding search
	return []SimilarChunk{}, nil
}

func (repo *PostgresRepoChunksRepository) GetRepoChunksCount(ctx context.Context, repoHash string) (int, error) {
	var count int
	err := repo.db.QueryRow(ctx, "SELECT COUNT(*) FROM repo_chunks WHERE repo_hash = $1", repoHash).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count repo chunks: %w", err)
	}
	return count, nil
}

func (repo *PostgresRepoChunksRepository) DeleteRepoChunks(ctx context.Context, repoHash string) error {
	_, err := repo.db.Exec(ctx, "DELETE FROM repo_chunks WHERE repo_hash = $1", repoHash)
	if err != nil {
		return fmt.Errorf("failed to delete repo chunks: %w", err)
	}
	return nil
}
