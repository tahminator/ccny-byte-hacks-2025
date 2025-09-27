package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	gogit "github.com/go-git/go-git/v5"
	githttp "github.com/go-git/go-git/v5/plumbing/transport/http"
	gh "github.com/google/go-github/v75/github"

	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
	"github.com/tahminator/go-react-template/utils"
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

	// --- POST /github/connect
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
		appUser := aoPtr.User

		// Validate the token and read the login
		cli := gh.NewClient(nil).WithAuthToken(body.Token)
		ghUser, resp, err := cli.Users.Get(c.Request.Context(), "")
		if err != nil || ghUser == nil || ghUser.Login == nil || *ghUser.Login == "" {
			c.JSON(http.StatusBadGateway, gin.H{"error": "failed to validate token with GitHub"})
			return
		}
		login := *ghUser.Login

		// Persist creds
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
		_ = resp

		c.JSON(http.StatusOK, gin.H{
			"message":         "github connected",
			"github_username": login,
			"token":           masked,
		})
	})

	// --- GET /github/repos
	r.GET("/repos", func(c *gin.Context) {
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
		userID := ao.User.Id

		u, err := userRepository.GetUserById(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if u.GithubToken == nil || *u.GithubToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "user does not have a GitHub token"})
			return
		}

		client := gh.NewClient(nil).WithAuthToken(*u.GithubToken)
		opt := &gh.RepositoryListByAuthenticatedUserOptions{
			Type:        "all",
			ListOptions: gh.ListOptions{PerPage: 100},
		}

		var repoNames []string
		ctx := c.Request.Context()
		for {
			repos, resp, err := client.Repositories.ListByAuthenticatedUser(ctx, opt)
			if err != nil {
				c.JSON(http.StatusBadGateway, gin.H{"error": "failed to fetch repos from GitHub"})
				return
			}
			for _, rpo := range repos {
				if rpo.Name != nil {
					repoNames = append(repoNames, *rpo.Name)
				}
			}
			if resp == nil || resp.NextPage == 0 {
				break
			}
			opt.Page = resp.NextPage
		}

		c.JSON(http.StatusOK, repoNames)
	})

	// --- POST /github/clone
	r.POST("/clone", func(c *gin.Context) {
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
		userID := ao.User.Id

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

	// --- POST /github/commit
	r.POST("/commit", func(c *gin.Context) {
		type Req struct {
			RepoName    string `json:"repoName"`
			NewFileData string `json:"newFileData"`
			Path        string `json:"path"`
		}

		var body Req

		if err := c.ShouldBindJSON(&body); err != nil ||
			strings.TrimSpace(body.RepoName) == "" || strings.TrimSpace(body.NewFileData) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repo name and/or new file data should not be empty"})
			return
		}

		ao := c.MustGet("ao").(*utils.AuthenticationObject)
		githubUsername := ao.User.GithubUsername
		githubToken := ao.User.GithubToken
		userId := ao.User.Id.String()

		base := filepath.Join("repos", userId, *githubUsername, body.RepoName)
		path := body.Path
		if len(path) > 0 && path[0] == '/' {
			path = path[1:]
		}
		fullPath := filepath.Join(base, path)

		if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, utils.Failure("failed to create parent directories"))
			return
		}
		if err := os.WriteFile(fullPath, []byte(body.NewFileData), 0o644); err != nil {
			c.JSON(http.StatusInternalServerError, utils.Failure("failed to write file"))
			return
		}

		status, stdout, stderr, err := utils.RunCommand(fmt.Sprintf(
			"cd %s && [ -f \"$(git rev-parse --git-dir)/MERGE_HEAD\" ] && echo true || echo false", base))
		if err != nil || stderr != "" || status != 0 {
			c.JSON(http.StatusInternalServerError, utils.Failure("failed to commit repository"))
			return
		}

		inMergeMode, err := strconv.ParseBool(strings.TrimSpace(stdout))
		if err != nil {
			c.JSON(http.StatusInternalServerError, utils.Failure("failed to commit repository"))
			return
		}

		if inMergeMode {
			status, stdout, stderr, err := utils.RunCommand(fmt.Sprintf("cd %s && git diff --name-only --diff-filter=U", base))
			if err != nil || stderr != "" || status != 0 {
				c.JSON(http.StatusInternalServerError, utils.Failure("failed to commit repository"))
				return
			}
			lines := strings.Split(strings.TrimSpace(stdout), "\n")
			if len(lines) > 0 && lines[0] != "" {
				c.JSON(http.StatusInternalServerError, utils.Failure("failed to commit repository"))
				return
			}
			url := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", *githubUsername, *githubToken, *githubUsername, body.RepoName)
			status, _, _, err = utils.RunCommand(fmt.Sprintf("cd %s && (git commit -m 'Test fix' || true) && git push %s", base, url))
			if status != 0 || err != nil {
				c.JSON(http.StatusInternalServerError, utils.Failure("failed to commit repository"))
				return
			}
			c.JSON(http.StatusOK, utils.Success("ok", gin.H{}))
			return
		}

		// Normal path
		status, _, _, err = utils.RunCommand(fmt.Sprintf("cd %s && git fetch && git merge", base))
		if status != 0 || err != nil {
			c.JSON(http.StatusInternalServerError, utils.Failure("failed to commit repository"))
			return
		}
		url := fmt.Sprintf("https://%s:%s@github.com/%s/%s.git", *githubUsername, *githubToken, *githubUsername, body.RepoName)
		status, _, _, err = utils.RunCommand(fmt.Sprintf("cd %s && git add . && (git commit -m 'Test msg' || true) && git push %s", base, url))
		if status != 0 || err != nil {
			c.JSON(http.StatusInternalServerError, utils.Failure("failed to commit repository"))
			return
		}
		c.JSON(http.StatusOK, utils.Success("ok", gin.H{}))
	})

	// --- POST /github/merge/accept
	r.POST("/merge/accept", func(c *gin.Context) {
		type req struct {
			NewFileData string `json:"newFileData"`
			FullPath    string `json:"fullPath"`
			RepoName    string `json:"repoName"`
		}

		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
			return
		}
		body.FullPath = strings.TrimSpace(body.FullPath)
		body.RepoName = strings.TrimSpace(body.RepoName)

		if body.RepoName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repoName is required"})
			return
		}
		if err := validateSlug(body.RepoName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repoName"})
			return
		}
		if body.FullPath == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fullPath is required"})
			return
		}
		if strings.HasPrefix(body.FullPath, "/") || strings.Contains(body.FullPath, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "fullPath must be a safe relative path"})
			return
		}

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

		u, err := userRepository.GetUserById(c.Request.Context(), ao.User.Id)
		if err != nil || u == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if u.GithubUsername == nil || strings.TrimSpace(*u.GithubUsername) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "github username not set for user"})
			return
		}
		owner := strings.TrimSpace(*u.GithubUsername)

		base := filepath.Join("repos", ao.User.Id.String(), owner)
		repoAbs := filepath.Join(base, body.RepoName)
		if st, err := os.Stat(repoAbs); err != nil || !st.IsDir() {
			c.JSON(http.StatusNotFound, gin.H{"error": "repo not found on disk"})
			return
		}

		relClean := filepath.Clean(body.FullPath)
		fileAbs := filepath.Join(repoAbs, relClean)

		repoAbsClean := filepath.Clean(repoAbs)
		fileAbsClean := filepath.Clean(fileAbs)
		sep := string(os.PathSeparator)
		if !strings.HasPrefix(fileAbsClean+sep, repoAbsClean+sep) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fullPath; escapes repo root"})
			return
		}

		if err := os.MkdirAll(filepath.Dir(fileAbsClean), 0o755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create parent directories"})
			return
		}
		if err := os.WriteFile(fileAbsClean, []byte(body.NewFileData), 0o644); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write file"})
			return
		}

		posixRel := filepath.ToSlash(relClean)
		if code, _, errOut, err := utils.RunCommand(fmt.Sprintf(`git -C %q add -- %q`, repoAbsClean, posixRel)); err != nil || code != 0 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "git add failed", "details": errOut})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "ok",
			"repoName": body.RepoName,
			"fullPath": posixRel,
			"staged":   true,
		})
	})

	// --- POST /github/merge/decline
	r.POST("/merge/decline", func(c *gin.Context) {
		type req struct {
			FullPath string `json:"fullPath"`
			RepoName string `json:"repoName"`
		}

		var body req
		if err := c.ShouldBindJSON(&body); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid json body"})
			return
		}
		body.RepoName = strings.TrimSpace(body.RepoName)
		if body.RepoName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "repoName is required"})
			return
		}
		if err := validateSlug(body.RepoName); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid repoName"})
			return
		}

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

		u, err := userRepository.GetUserById(c.Request.Context(), ao.User.Id)
		if err != nil || u == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		if u.GithubUsername == nil || strings.TrimSpace(*u.GithubUsername) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "github username not set for user"})
			return
		}
		owner := strings.TrimSpace(*u.GithubUsername)

		base := filepath.Join("repos", ao.User.Id.String(), owner, body.RepoName)
		if st, err := os.Stat(base); err != nil || !st.IsDir() {
			c.JSON(http.StatusNotFound, gin.H{"error": "repo not found on disk"})
			return
		}

		cmd := fmt.Sprintf("cd %s && git merge --abort && git reset --hard HEAD~1", base)
		status, stdout, stderr, err := utils.RunCommand(cmd)
		if err != nil || status != 0 {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to decline merge",
				"details": stderr,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "merge declined",
			"repoName": body.RepoName,
			"action":   "merge-abort-reset",
			"stdout":   stdout,
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
