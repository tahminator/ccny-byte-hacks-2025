package gemini

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tahminator/go-react-template/database/repository/repo_chunks"
	"google.golang.org/genai"
)

func NewRouter(eng *gin.RouterGroup,
	geminiClient *genai.Client,
	repoChunksRepo repo_chunks.RepoChunksRepository,
) *gin.RouterGroup {
	r := eng.Group("/gemini")

	service := NewGeminiService(geminiClient, repoChunksRepo)

	r.GET("/test", func(c *gin.Context) {
		message := c.Query("message")
		if message == "" {
			message = "solve two-sum for me"
		}

		StreamGeminiResponse(c, geminiClient, message)
	})

	r.POST("/resolve-conflicts-file-stream", func(c *gin.Context) {
		var req struct {
			ConflictContent string `json:"conflict_content" binding:"required"`
			FilePath        string `json:"file_path" binding:"required"`
			UserQuery       string `json:"user_query"`
			RepoHash        string `json:"repo_hash"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.UserQuery == "" {
			req.UserQuery = "resolve all merge conflicts in this code"
		}
		if req.RepoHash == "" {
			req.RepoHash = "default"
		}

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

		service.StreamResolveConflictsToFile(c, req.ConflictContent, req.FilePath, req.UserQuery, req.RepoHash)
	})

	return r
}
