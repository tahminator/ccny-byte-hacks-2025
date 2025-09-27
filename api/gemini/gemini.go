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

	r.POST("/resolve-conflicts", func(c *gin.Context) {
		var req struct {
			Query    string `json:"query" binding:"required"`
			RepoHash string `json:"repo_hash" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := service.ResolveMergeConflictsWithRAG(c.Request.Context(), req.Query, req.RepoHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"commands": response})
	})

	r.POST("/resolve-semantic", func(c *gin.Context) {
		var req struct {
			Query    string `json:"query" binding:"required"`
			RepoHash string `json:"repo_hash" binding:"required"`
			K        int    `json:"k"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.K == 0 {
			req.K = 10
		}

		response, err := service.ResolveConflictsWithSemanticSearch(c.Request.Context(), req.Query, req.RepoHash, req.K)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"commands": response})
	})

	r.POST("/resolve-threshold", func(c *gin.Context) {
		var req struct {
			Query       string  `json:"query" binding:"required"`
			RepoHash    string  `json:"repo_hash" binding:"required"`
			K           int     `json:"k"`
			MaxDistance float64 `json:"max_distance"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.K == 0 {
			req.K = 10
		}
		if req.MaxDistance == 0 {
			req.MaxDistance = 0.5
		}

		response, err := service.ResolveConflictsWithThreshold(c.Request.Context(), req.Query, req.RepoHash, req.K, req.MaxDistance)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"commands": response})
	})

	r.POST("/analyze-conflict", func(c *gin.Context) {
		var req struct {
			ConflictContent string `json:"conflict_content" binding:"required"`
			FilePath        string `json:"file_path" binding:"required"`
			Language        string `json:"language"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if req.Language == "" {
			req.Language = "unknown"
		}

		response, err := service.AnalyzeCodeConflict(c.Request.Context(), req.ConflictContent, req.FilePath, req.Language)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"commands": response})
	})

	r.POST("/generate-commands", func(c *gin.Context) {
		var req struct {
			Query         string   `json:"query" binding:"required"`
			ConflictFiles []string `json:"conflict_files"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := service.GenerateShellCommands(c.Request.Context(), req.Query, req.ConflictFiles)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"commands": response})
	})

	r.POST("/automated-resolution", func(c *gin.Context) {
		var req struct {
			ConflictFiles []string `json:"conflict_files" binding:"required"`
			RepoContext   string   `json:"repo_context"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := service.AutomatedResolution(c.Request.Context(), req.ConflictFiles, req.RepoContext)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"commands": response})
	})

	r.POST("/interactive-resolution", func(c *gin.Context) {
		var req struct {
			Query           string            `json:"query" binding:"required"`
			ConflictDetails map[string]string `json:"conflict_details" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		response, err := service.InteractiveResolution(c.Request.Context(), req.Query, req.ConflictDetails)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"commands": response})
	})

	r.GET("/conflict-files/:repo_hash", func(c *gin.Context) {
		repoHash := c.Param("repo_hash")
		if repoHash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repo_hash is required"})
			return
		}

		files, err := service.GetConflictFiles(c.Request.Context(), repoHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"files": files})
	})

	r.GET("/repo-context/:repo_hash", func(c *gin.Context) {
		repoHash := c.Param("repo_hash")
		if repoHash == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repo_hash is required"})
			return
		}

		context, err := service.GetRepoContext(c.Request.Context(), repoHash)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"context": context})
	})

	r.POST("/validate-commands", func(c *gin.Context) {
		var req struct {
			Commands string `json:"commands" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validatedCommands, err := service.ValidateShellCommands(req.Commands)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"valid_commands": validatedCommands})
	})

	return r
}
