package gemini

import (
	"context"
	"fmt"
	"strings"

	"github.com/tahminator/go-react-template/database/repository/repo_chunks"
	"google.golang.org/genai"
)

type GeminiService struct {
	client         *genai.Client
	repoChunksRepo repo_chunks.RepoChunksRepository
}

func NewGeminiService(client *genai.Client, repoChunksRepo repo_chunks.RepoChunksRepository) *GeminiService {
	return &GeminiService{
		client:         client,
		repoChunksRepo: repoChunksRepo,
	}
}

func (gs *GeminiService) ResolveMergeConflictsWithRAG(
	ctx context.Context,
	userQuery string,
	repoHash string,
) (string, error) {
	conflictChunks, err := gs.repoChunksRepo.GetConflictChunks(ctx, repoHash)
	if err != nil {
		return "", fmt.Errorf("failed to get conflict chunks: %w", err)
	}

	var contextChunks []repo_chunks.SimilarChunk
	if len(conflictChunks) > 0 {
		for _, conflict := range conflictChunks {
			ctxChunks, err := gs.repoChunksRepo.GetContextChunks(ctx, repoHash, conflict.Source, conflict.LineStart, conflict.LineEnd, 10)
			if err != nil {
				continue
			}
			contextChunks = append(contextChunks, ctxChunks...)
		}
	}

	similarFunctions, err := gs.repoChunksRepo.GetSimilarFunctions(ctx, userQuery, repoHash, 5)
	if err != nil {
		similarFunctions = []repo_chunks.SimilarChunk{}
	}

	prompt := Prompt + "\n\nUser Request: " + userQuery + "\n\n"

	if len(conflictChunks) > 0 {
		prompt += "MERGE CONFLICTS DETECTED:\n"
		for i, chunk := range conflictChunks {
			prompt += fmt.Sprintf("Conflict %d in %s (lines %d-%d):\n", i+1, chunk.Source, chunk.LineStart, chunk.LineEnd)
			prompt += fmt.Sprintf("Section: %s\n", chunk.ConflictSection)
			prompt += fmt.Sprintf("Content:\n%s\n\n", chunk.Chunk)
		}
	}

	if len(contextChunks) > 0 {
		prompt += "RELEVANT CONTEXT:\n"
		for i, chunk := range contextChunks {
			prompt += fmt.Sprintf("Context %d from %s:\n%s\n\n", i+1, chunk.Source, chunk.Chunk)
		}
	}

	if len(similarFunctions) > 0 {
		prompt += "SIMILAR FUNCTIONS FOR REFERENCE:\n"
		for i, chunk := range similarFunctions {
			prompt += fmt.Sprintf("Function %d from %s:\n%s\n\n", i+1, chunk.Source, chunk.Chunk)
		}
	}

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) ResolveConflictsToFile(
	ctx context.Context,
	conflictContent string,
	filePath string,
	userQuery string,
	repoHash string,
) (string, error) {
	similarChunks, err := gs.repoChunksRepo.GetSimilarChunks(ctx, userQuery, repoHash, 5)
	if err != nil {
		similarChunks = []repo_chunks.SimilarChunk{}
	}

	prompt := Prompt + "\n\nUser Request: " + userQuery + "\n\n"
	prompt += fmt.Sprintf("File: %s\n", filePath)
	prompt += "CONFLICTED FILE CONTENT:\n"
	prompt += conflictContent + "\n\n"

	if len(similarChunks) > 0 {
		prompt += "REPOSITORY CONTEXT:\n"
		for i, chunk := range similarChunks {
			prompt += fmt.Sprintf("Context %d from %s (similarity: %.3f):\n", i+1, chunk.Source, chunk.Distance)
			prompt += fmt.Sprintf("Content:\n%s\n\n", chunk.Chunk)
		}
	}

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate resolved file: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) ResolveConflictsWithSemanticSearch(
	ctx context.Context,
	userQuery string,
	repoHash string,
	k int,
) (string, error) {
	similarChunks, err := gs.repoChunksRepo.GetSimilarChunks(ctx, userQuery, repoHash, k)
	if err != nil {
		return "", fmt.Errorf("failed to get similar chunks: %w", err)
	}

	prompt := Prompt + "\n\nUser Request: " + userQuery + "\n\n"

	if len(similarChunks) > 0 {
		prompt += "REPOSITORY CONTEXT (from semantic search):\n"
		for i, chunk := range similarChunks {
			prompt += fmt.Sprintf("Context %d from %s (similarity: %.3f):\n", i+1, chunk.Source, chunk.Distance)
			if chunk.FileType == "conflict" {
				prompt += fmt.Sprintf("Conflict section: %s\n", chunk.ConflictSection)
			}
			prompt += fmt.Sprintf("Content:\n%s\n\n", chunk.Chunk)
		}
	}

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) ResolveConflictsWithThreshold(
	ctx context.Context,
	userQuery string,
	repoHash string,
	k int,
	maxDistance float64,
) (string, error) {
	similarChunks, err := gs.repoChunksRepo.GetSimilarChunksWithThreshold(ctx, userQuery, repoHash, k, maxDistance)
	if err != nil {
		return "", fmt.Errorf("failed to get similar chunks with threshold: %w", err)
	}

	prompt := Prompt + "\n\nUser Request: " + userQuery + "\n\n"

	if len(similarChunks) > 0 {
		prompt += "REPOSITORY CONTEXT (filtered by similarity threshold):\n"
		for i, chunk := range similarChunks {
			prompt += fmt.Sprintf("Context %d from %s (similarity: %.3f):\n", i+1, chunk.Source, chunk.Distance)
			if chunk.FileType == "conflict" {
				prompt += fmt.Sprintf("Conflict section: %s\n", chunk.ConflictSection)
			}
			prompt += fmt.Sprintf("Content:\n%s\n\n", chunk.Chunk)
		}
	}

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate response: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) AnalyzeCodeConflict(
	ctx context.Context,
	conflictContent string,
	filePath string,
	language string,
) (string, error) {
	prompt := Prompt + "\n\nFile: " + filePath + "\nLanguage: " + language + "\n\nConflict content:\n" + conflictContent + "\n\n"

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to analyze code conflict: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) GenerateShellCommands(
	ctx context.Context,
	userQuery string,
	conflictFiles []string,
) (string, error) {
	prompt := Prompt + "\n\nTask: " + userQuery + "\n\n"

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate shell commands: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) AutomatedResolution(
	ctx context.Context,
	conflictFiles []string,
	repoContext string,
) (string, error) {
	prompt := Prompt + "\n\n"

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate automated resolution: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) InteractiveResolution(
	ctx context.Context,
	userQuery string,
	conflictDetails map[string]string,
) (string, error) {
	prompt := Prompt + "\n\nUser request: " + userQuery + "\n\n"

	response, err := gs.generateResponse(ctx, prompt)
	if err != nil {
		return "", fmt.Errorf("failed to generate interactive resolution: %w", err)
	}

	return response, nil
}

func (gs *GeminiService) generateResponse(ctx context.Context, prompt string) (string, error) {
	thinkingBudget := int32(0)
	iter := gs.client.Models.GenerateContentStream(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		&genai.GenerateContentConfig{
			ThinkingConfig: &genai.ThinkingConfig{
				ThinkingBudget: &thinkingBudget,
			},
		},
	)

	var response strings.Builder
	for resp, err := range iter {
		if err != nil {
			return "", fmt.Errorf("streaming error: %w", err)
		}

		if resp != nil && resp.Text() != "" {
			response.WriteString(resp.Text())
		}
	}

	return response.String(), nil
}

func (gs *GeminiService) ValidateShellCommands(response string) ([]string, error) {
	lines := strings.Split(strings.TrimSpace(response), "\n")
	var commands []string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if gs.isValidShellCommand(line) {
			commands = append(commands, line)
		} else {
			return nil, fmt.Errorf("invalid shell command: %s", line)
		}
	}

	return commands, nil
}

func (gs *GeminiService) isValidShellCommand(line string) bool {
	line = strings.TrimPrefix(line, "$")
	line = strings.TrimPrefix(line, "#")

	validPrefixes := []string{
		"git", "sed", "awk", "grep", "find", "ls", "cat", "echo",
		"mv", "cp", "rm", "mkdir", "touch", "chmod", "chown",
		"vim", "nano", "emacs", "vi", "nano", "code", "subl",
		"curl", "wget", "tar", "unzip", "zip", "gzip",
		"ps", "kill", "top", "htop", "df", "du", "free",
		"cd", "pwd", "export", "unset", "alias", "unalias",
		"source", ".", "exec", "eval", "test", "[", "[[",
		"if", "then", "else", "elif", "fi", "for", "while", "do", "done",
		"case", "esac", "function", "return", "exit",
	}

	line = strings.TrimSpace(line)
	for _, prefix := range validPrefixes {
		if strings.HasPrefix(line, prefix+" ") || line == prefix {
			return true
		}
	}

	if strings.Contains(line, "=") && !strings.Contains(line, " ") {
		return true
	}

	if strings.HasPrefix(line, "#") {
		return false
	}

	if strings.Contains(line, "This") || strings.Contains(line, "The") ||
		strings.Contains(line, "You") || strings.Contains(line, "We") ||
		strings.Contains(line, "Note:") || strings.Contains(line, "Warning:") ||
		strings.Contains(line, "Error:") || strings.Contains(line, "Info:") {
		return false
	}

	return false
}

func (gs *GeminiService) GetConflictFiles(ctx context.Context, repoHash string) ([]string, error) {
	conflictChunks, err := gs.repoChunksRepo.GetConflictChunks(ctx, repoHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get conflict chunks: %w", err)
	}

	fileMap := make(map[string]bool)
	for _, chunk := range conflictChunks {
		fileMap[chunk.Source] = true
	}

	var files []string
	for file := range fileMap {
		files = append(files, file)
	}

	return files, nil
}

func (gs *GeminiService) GetRepoContext(ctx context.Context, repoHash string) (string, error) {
	count, err := gs.repoChunksRepo.GetRepoChunksCount(ctx, repoHash)
	if err != nil {
		return "", fmt.Errorf("failed to get repo chunks count: %w", err)
	}

	sampleChunks, err := gs.repoChunksRepo.GetSimilarChunks(ctx, "repository structure", repoHash, 10)
	if err != nil {
		return "", fmt.Errorf("failed to get sample chunks: %w", err)
	}

	var context strings.Builder
	context.WriteString(fmt.Sprintf("Repository has %d chunks\n", count))
	context.WriteString("Sample files and structure:\n")

	for _, chunk := range sampleChunks {
		context.WriteString(fmt.Sprintf("- %s (%s)\n", chunk.Source, chunk.ChunkType))
	}

	return context.String(), nil
}
