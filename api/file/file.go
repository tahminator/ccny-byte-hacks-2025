package file

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
	"github.com/tahminator/go-react-template/utils"
)

func NewRouter(eng *gin.RouterGroup,
	userRepository user.UserRepository,
	sessionRepository session.SessionRepository,
) *gin.RouterGroup {
	r := eng.Group("/file")

	r.Use(func(c *gin.Context) {
		ao, err := utils.ValidateRequest(c, userRepository, sessionRepository)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		c.Set("ao", ao)
		c.Next()
	})

	r.GET("/data/*path", func(c *gin.Context) {
		ao := c.MustGet("ao").(*utils.AuthenticationObject)

		relPath := c.Param("path")
		if len(relPath) > 0 && relPath[0] == '/' {
			relPath = relPath[1:]
		}

		base := filepath.Join("repos", ao.User.Id.String())
		fullPath := filepath.Join(base, relPath)

		data, err := os.ReadFile(fullPath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		c.Data(http.StatusOK, "text/plain; charset=utf-8", data)
	})

	r.POST("/data/*path", func(c *gin.Context) {
		type Req struct {
			Content string `json:"content"`
		}

		var body Req
		if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Content) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "content should not be empty"})
			return
		}
		ao := c.MustGet("ao").(*utils.AuthenticationObject)

		relPath := c.Param("path")
		if len(relPath) > 0 && relPath[0] == '/' {
			relPath = relPath[1:]
		}

		base := filepath.Join("repos", ao.User.Id.String())
		fullPath := filepath.Join(base, relPath)
		permissions := os.FileMode(0o644)

		err := os.WriteFile(fullPath, []byte(body.Content), permissions)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		c.JSON(http.StatusOK, utils.Success("ok", gin.H{}))
	})

	return r
}
