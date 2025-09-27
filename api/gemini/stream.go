package gemini

import (
	"log"

	"github.com/gin-gonic/gin"
	"google.golang.org/genai"
)

func StreamGeminiResponse(c *gin.Context, geminiClient *genai.Client, message string) { // if error, close the stream and print the error in the backend
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

	c.Header("Content-Type", "text/plain")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	c.Writer.Flush()

	for resp, err := range iter {
		if err != nil {
			log.Printf("Gemini streaming error: %v", err)
			c.Writer.WriteString("Error: " + err.Error())
			c.Writer.Flush()
			return
		}

		if resp != nil && resp.Text() != "" {
			chunk := resp.Text()
			c.Writer.WriteString(chunk)
			c.Writer.Flush()
		}
	}

	c.Writer.Flush()
}
