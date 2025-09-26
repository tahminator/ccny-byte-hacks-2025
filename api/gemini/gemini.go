package gemini

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

func NewRouter(eng *gin.RouterGroup,
	geminiClient *genai.Client,
) *gin.RouterGroup {
	r := eng.Group("/gemini")

	r.GET("/test", func(c *gin.Context) {
		message := c.Query("message")
		if message == "" {
			message = "solve two-sum for me"
		}

		StreamGeminiResponse(c, geminiClient, message)
	})

	return r
}
