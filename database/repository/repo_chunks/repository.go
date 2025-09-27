package repo_chunks

import (
	"context"
)

type RepoChunksRepository interface {
	InsertChunk(ctx context.Context, chunk *RepoChunk) error
	GetConflictChunks(ctx context.Context, repoHash string) ([]SimilarChunk, error)
	GetContextChunks(ctx context.Context, repoHash, source string, lineStart, lineEnd, contextLines int) ([]SimilarChunk, error)
	GetSimilarChunks(ctx context.Context, query, repoHash string, k int) ([]SimilarChunk, error)
	GetSimilarChunksWithThreshold(ctx context.Context, query, repoHash string, k int, maxDistance float64) ([]SimilarChunk, error)
	GetSimilarFunctions(ctx context.Context, query, repoHash string, k int) ([]SimilarChunk, error)
	GetRepoChunksCount(ctx context.Context, repoHash string) (int, error)
	DeleteRepoChunks(ctx context.Context, repoHash string) error
}
