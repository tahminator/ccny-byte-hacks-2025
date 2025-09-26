package gemini

import (
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

func StreamGeminiResponse(c *gin.Context, geminiClient *genai.Client, message string) {     // if error, close the stream and print the error in the backend
	thinkingBudget := int32(0)
	iter := geminiClient.Models.GenerateContentStream(
		c.Request.Context(),
		"gemini-2.5-flash",
		genai.Text(message),
		&genai.GenerateContentConfig{
			ThinkingConfig: &genai.ThinkingConfig{
				ThinkingBudget: &thinkingBudget,
			},
		},
	)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	c.SSEvent("start", gin.H{
		"query":  message,
		"status": "streaming",
	})
	c.Writer.Flush()

	var buffer string

	for resp, err := range iter {
		if err != nil {
			log.Printf("Gemini streaming error: %v", err)

			c.SSEvent("error", gin.H{
				"message": "Streaming error occurred",
				"error":   err.Error(),
				"status":  "error",
			})
			c.Writer.Flush()
			return 
		}

		if resp != nil && resp.Text() != "" {
			chunk := resp.Text()
			buffer += chunk

			c.SSEvent("delta", gin.H{
				"content": chunk,
				"buffer":  buffer,
			})
			c.Writer.Flush()
		}
	}

	c.SSEvent("done", gin.H{
		"status":       "complete",
		"finalContent": buffer,
	})
	c.Writer.Flush()
}

