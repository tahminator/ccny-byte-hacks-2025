package github

import (
	"github.com/google/go-github/v75/github"
	"github.com/google/uuid"
	"github.com/tahminator/go-react-template/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
)

func NewRouter(eng *gin.RouterGroup, userRepository user.UserRepository, sessionRepository session.SessionRepository) *gin.RouterGroup {
	r := eng.Group("/github")

	r.Use(func(c *gin.Context) {
		ao, err := utils.ValidateRequest(c, userRepository, sessionRepository)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}
		c.Set("userID", ao.User.Id.String())
		c.Next()
	})

	r.GET("/repos", func(context *gin.Context) {
		userIDStr, ok := context.Get("userID")
		if !ok {
			context.JSON(http.StatusUnauthorized, gin.H{"error": "missing user in context"})
			return
		}
		uidStr, _ := userIDStr.(string)
		userID, err := uuid.Parse(uidStr)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}

		u, err := userRepository.GetUserById(context.Request.Context(), userID)
		if err != nil {
			context.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		if u.GithubToken == nil || *u.GithubToken == "" {
			context.JSON(http.StatusBadRequest, gin.H{"error": "user does not have a GitHub token"})
			return
		}
		ghToken := u.GithubToken

		client := github.NewClient(nil).WithAuthToken(*ghToken)

		opt := &github.RepositoryListByAuthenticatedUserOptions{
			Type:        "all",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		var repoNames []string
		ctx := context.Request.Context()
		for {
			repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
			if err != nil {
				context.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch repos from GitHub"})
				return
			}
			for _, r := range repos {
				if r.Name != nil {
					repoNames = append(repoNames, *r.Name)
				}
			}
			if resp == nil || resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		context.JSON(http.StatusOK, repoNames)
	})
	return r
}
