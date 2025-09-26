package rag

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/tahminator/go-react-template/database"
	"google.golang.org/genai"
)

const (
	EmbedModel   = "text-embedding-004"
	EmbedDim     = 768
	ChunkSize    = 300
	ChunkOverlap = 30
	BatchSize    = 50
)

type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
}

type Chunk struct {
	Content         string
	Source          string
	FileType        string
	ConflictSection string
	LineStart       int
	LineEnd         int
	ChunkType       string
}

type RepoChunk struct {
	RepoHash  string    `db:"repo_hash"`
	Source    string    `db:"source"`
	Chunk     string    `db:"chunk"`
	Embedding []float64 `db:"embedding"`
}

func RepoHash(files []FileContent) string {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})

	fileData := make([]map[string]string, len(files))
	for i, f := range files {
		fileData[i] = map[string]string{
			"path":    f.Path,
			"content": f.Content,
		}
	}

	data, _ := json.Marshal(fileData)
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func splitTextIntoChunks(text string, chunkSize, overlap int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0

	for start < len(text) {
		end := start + chunkSize
		if end > len(text) {
			end = len(text)
		}

		if end < len(text) {
			lastSpace := strings.LastIndex(text[start:end], " ")
			if lastSpace > 0 {
				end = start + lastSpace
			}
		}

		chunk := text[start:end]
		chunks = append(chunks, chunk)

		start = end - overlap
		if start >= len(text) {
			break
		}
	}

	return chunks
}

func detectConflictMarkers(text string) (bool, string, []int) {
	lines := strings.Split(text, "\n")
	var conflictLines []int
	hasConflicts := false
	conflictType := ""

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "<<<<<<<") {
			hasConflicts = true
			conflictType = "ours"
			conflictLines = append(conflictLines, i)
		} else if strings.HasPrefix(line, "=======") {
			conflictType = "theirs"
			conflictLines = append(conflictLines, i)
		} else if strings.HasPrefix(line, ">>>>>>>") {
			conflictType = "end"
			conflictLines = append(conflictLines, i)
		}
	}

	return hasConflicts, conflictType, conflictLines
}

func analyzeCodeChunk(content string) string {
	content = strings.TrimSpace(content)

	if strings.Contains(content, "<<<<<<<") || strings.Contains(content, "=======") || strings.Contains(content, ">>>>>>>") {
		return "conflict_marker"
	}

	if strings.HasPrefix(content, "import ") || strings.HasPrefix(content, "package ") {
		return "import"
	}

	if strings.Contains(content, "func ") && strings.Contains(content, "(") {
		return "function"
	}

	if strings.HasPrefix(content, "//") || strings.HasPrefix(content, "/*") {
		return "comment"
	}

	return "code"
}

func createChunksFromFiles(files []FileContent) []Chunk {
	var chunks []Chunk

	for _, file := range files {
		hasConflicts, _, _ := detectConflictMarkers(file.Content)

		if hasConflicts {
			chunks = append(chunks, createConflictChunks(file)...)
		} else {
			fileChunks := splitTextIntoChunks(file.Content, ChunkSize, ChunkOverlap)
			for i, chunk := range fileChunks {
				chunkType := analyzeCodeChunk(chunk)
				chunks = append(chunks, Chunk{
					Content:   chunk,
					Source:    file.Path,
					FileType:  "context",
					LineStart: i * (ChunkSize - ChunkOverlap),
					LineEnd:   (i+1)*(ChunkSize-ChunkOverlap) + ChunkOverlap,
					ChunkType: chunkType,
				})
			}
		}
	}

	return chunks
}

func createConflictChunks(file FileContent) []Chunk {
	var chunks []Chunk
	lines := strings.Split(file.Content, "\n")

	var currentSection []string
	var sectionType string
	var lineStart int

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "<<<<<<<") {

			if len(currentSection) > 0 {
				chunks = append(chunks, createChunkFromSection(currentSection, file.Path, sectionType, lineStart, i-1))
			}
			currentSection = []string{line}
			sectionType = "ours"
			lineStart = i
		} else if strings.HasPrefix(trimmed, "=======") {

			if len(currentSection) > 0 {
				chunks = append(chunks, createChunkFromSection(currentSection, file.Path, sectionType, lineStart, i-1))
			}
			currentSection = []string{line}
			sectionType = "theirs"
			lineStart = i
		} else if strings.HasPrefix(trimmed, ">>>>>>>") {

			if len(currentSection) > 0 {
				chunks = append(chunks, createChunkFromSection(currentSection, file.Path, sectionType, lineStart, i-1))
			}
			currentSection = []string{line}
			sectionType = "end"
			lineStart = i
		} else {
			currentSection = append(currentSection, line)
		}
	}

	if len(currentSection) > 0 {
		chunks = append(chunks, createChunkFromSection(currentSection, file.Path, sectionType, lineStart, len(lines)-1))
	}

	return chunks
}

func createChunkFromSection(section []string, source, sectionType string, lineStart, lineEnd int) Chunk {
	content := strings.Join(section, "\n")
	chunkType := analyzeCodeChunk(content)

	return Chunk{
		Content:         content,
		Source:          source,
		FileType:        "conflict",
		ConflictSection: sectionType,
		LineStart:       lineStart,
		LineEnd:         lineEnd,
		ChunkType:       chunkType,
	}
}

func getEmbeddings(ctx context.Context, texts []string) ([][]float64, error) {
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

	var allEmbeddings [][]float64

	for i := 0; i < len(texts); i += BatchSize {
		end := i + BatchSize
		if end > len(texts) {
			end = len(texts)
		}

		batchTexts := texts[i:end]
		contents := make([]*genai.Content, len(batchTexts))
		for j, text := range batchTexts {
			contents[j] = genai.NewContentFromText(text, genai.RoleUser)
		}

		embedding, err := client.Models.EmbedContent(ctx, EmbedModel, contents, &genai.EmbedContentConfig{
			TaskType: "RETRIEVAL_DOCUMENT",
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get embeddings for batch: %w", err)
		}

		for _, emb := range embedding.Embeddings {
			embeddingValues := make([]float64, len(emb.Values))
			for j, val := range emb.Values {
				embeddingValues[j] = float64(val)
			}
			allEmbeddings = append(allEmbeddings, embeddingValues)
		}
	}

	return allEmbeddings, nil
}

func EmbedRepoPgVector(ctx context.Context, fileContents []FileContent) (string, error) {
	pool, err := database.GetPool()
	if err != nil {
		return "", fmt.Errorf("failed to get database pool: %w", err)
	}

	repoHash := RepoHash(fileContents)
	var exists bool
	err = pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM repo_chunks WHERE repo_hash = $1 LIMIT 1)", repoHash).Scan(&exists)
	if err != nil {
		return "", fmt.Errorf("failed to check if repo hash exists: %w", err)
	}

	if exists {
		return fmt.Sprintf("Using cached pgvector embeddings for %d files.", len(fileContents)), nil
	}

	chunks := createChunksFromFiles(fileContents)
	if len(chunks) == 0 {
		return "No content to embed", nil
	}

	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	embeddings, err := getEmbeddings(ctx, texts)
	if err != nil {
		return "", fmt.Errorf("failed to get embeddings: %w", err)
	}

	// Insert chunks one by one to avoid connection issues
	for i, chunk := range chunks {
		// Convert embedding to pgvector format
		embeddingValues := make([]string, len(embeddings[i]))
		for j, val := range embeddings[i] {
			embeddingValues[j] = fmt.Sprintf("%.6f", val)
		}
		embeddingStr := "[" + strings.Join(embeddingValues, ",") + "]"

		_, err := pool.Exec(ctx, "INSERT INTO repo_chunks (repo_hash, source, chunk, embedding, file_type, conflict_section, line_start, line_end, chunk_type) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)",
			repoHash, chunk.Source, chunk.Content, embeddingStr, chunk.FileType, chunk.ConflictSection, chunk.LineStart, chunk.LineEnd, chunk.ChunkType)
		if err != nil {
			return "", fmt.Errorf("failed to insert chunk %d: %w", i, err)
		}
	}

	return fmt.Sprintf("Embedded %d chunks from %d files into Postgres.", len(chunks), len(fileContents)), nil
}
