package github

import (
	"context"
	"errors"
	"github.com/google/go-github/v75/github"
	_ "github.com/google/uuid"
	"github.com/tahminator/go-react-template/utils"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	_ "regexp"
	"strings"
	_ "strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"

	gogit "github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gh "github.com/google/go-github/v75/github"
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
		c.Set("ao", ao)
		c.Next()
	})
	r.POST("/connect", func(c *gin.Context) {
		type req struct {
			Token string `json:"token"`
		}

		var body req
		if err := c.ShouldBindJSON(&body); err != nil || strings.TrimSpace(body.Token) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
			return
		}
		if len(body.Token) < 20 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token looks invalid"})
			return
		}

		aoRaw, ok := c.Get("ao")
		if !ok || aoRaw == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user in context"})
			return
		}
		var aoPtr *utils.AuthenticationObject
		if v, ok := aoRaw.(*utils.AuthenticationObject); ok && v != nil {
			aoPtr = v
		} else if v2, ok := aoRaw.(utils.AuthenticationObject); ok {
			aoPtr = &v2
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth object in context"})
			return
		}
		appUser := aoPtr.User // your validated, loaded app user

		// 2) Validate the token with GitHub and read the login
		cli := gh.NewClient(nil).WithAuthToken(body.Token)
		ghUser, resp, err := cli.Users.Get(c.Request.Context(), "")
		if err != nil || ghUser == nil || ghUser.Login == nil || *ghUser.Login == "" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to validate token with GitHub"})
			return
		}
		login := *ghUser.Login

		// 3) Persist creds (fields are *string on your user model)
		token := body.Token
		appUser.GithubToken = &token
		appUser.GithubUsername = &login

		if _, err := userRepository.UpdateUser(c.Request.Context(), appUser); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save github creds"})
			return
		}

		masked := token
		if len(masked) > 8 {
			masked = masked[:4] + "****" + masked[len(masked)-4:]
		}

		_ = resp // kept to avoid “unused”

		c.JSON(http.StatusOK, gin.H{
			"message":         "github connected",
			"github_username": login,
			"token":           masked,
		})
	})

	r.GET("/repos", func(c *gin.Context) {
		aoVal, ok := c.Get("ao")
		if !ok || aoVal == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth object in context"})
			return
		}
		ao, ok := aoVal.(*utils.AuthenticationObject) // adjust to your actual type if different
		if !ok || ao == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth object in context"})
			return
		}
		userID := ao.User.Id // uuid.UUID

		u, err := userRepository.GetUserById(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if u.GithubToken == nil || *u.GithubToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user does not have a GitHub token"})
			return
		}

		client := github.NewClient(nil).WithAuthToken(*u.GithubToken)
		opt := &github.RepositoryListByAuthenticatedUserOptions{
			Type:        "all",
			ListOptions: github.ListOptions{PerPage: 100},
		}

		var repoNames []string
		ctx := c.Request.Context()
		for {
			repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch repos from GitHub"})
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

		c.JSON(http.StatusOK, repoNames)
	})

	r.POST("/clone", func(c *gin.Context) {
		// pull auth object from context
		aoVal, ok := c.Get("ao")
		if !ok || aoVal == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth object in context"})
			return
		}
		ao, ok := aoVal.(*utils.AuthenticationObject)
		if !ok || ao == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid auth object in context"})
			return
		}
		userID := ao.User.Id // uuid.UUID

		var body struct {
			Owner string `json:"owner"`
			Repo  string `json:"repo"`
			Force bool   `json:"force"`
		}
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
			return
		}
		if err := validateSlug(body.Owner); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid owner"})
			return
		}
		if err := validateSlug(body.Repo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repo"})
			return
		}

		u, err := userRepository.GetUserById(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if u.GithubToken == nil || *u.GithubToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user does not have a GitHub token"})
			return
		}
		token := *u.GithubToken

		ghClient := gh.NewClient(nil).WithAuthToken(token)
		repoInfo, _, err := ghClient.Repositories.Get(c.Request.Context(), body.Owner, body.Repo)
		defaultBranch := ""
		if err == nil && repoInfo != nil && repoInfo.GetDefaultBranch() != "" {
			defaultBranch = repoInfo.GetDefaultBranch()
		}

		destPath := filepath.Join("repos", userID.String(), body.Owner, body.Repo)

		if _, statErr := os.Stat(destPath); statErr == nil {
			if !body.Force {
				c.JSON(http.StatusConflict, gin.H{
					"error":       "destination already exists",
					"destination": destPath,
					"suggestion":  "use force=true to replace or delete it manually",
				})
				return
			}
			if err := os.RemoveAll(destPath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove existing destination"})
				return
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to access destination"})
			return
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create parent directories"})
			return
		}

		cloneURL := "https://github.com/" + body.Owner + "/" + body.Repo + ".git"

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Minute)
		defer cancel()

		_, err = gogit.PlainCloneContext(ctx, destPath, false, &gogit.CloneOptions{
			URL:   cloneURL,
			Depth: 1,
			Auth: &githttp.BasicAuth{
				Username: "git",
				Password: token,
			},
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"error":       "failed to clone repository",
				"details":     err.Error(),
				"destination": destPath,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":        "cloned",
			"owner":          body.Owner,
			"repo":           body.Repo,
			"default_branch": defaultBranch,
			"shallow":        true,
			"destination":    destPath,
		})
	})

	return r
}

func validateSlug(s string) error {
	re := regexp.MustCompile(`^[A-Za-z0-9._-]+$`)
	if !re.MatchString(s) {
		return errors.New("bad slug")
	}
	return nil
}
