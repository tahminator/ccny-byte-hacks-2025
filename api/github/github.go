package github

import (
	"context"
	"errors"
	"github.com/google/go-github/v75/github"
	"github.com/google/uuid"
	_ "github.com/google/uuid"
	"github.com/tahminator/go-react-template/utils"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	_ "regexp"
	"sort"
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

type CodeExtension string

const (
	ExtJS      CodeExtension = "JS"
	ExtTS      CodeExtension = "TS"
	ExtTSX     CodeExtension = "TSX"
	ExtJSX     CodeExtension = "JSX"
	ExtHTML    CodeExtension = "HTML"
	ExtCSS     CodeExtension = "CSS"
	ExtSCSS    CodeExtension = "SCSS"
	ExtMD      CodeExtension = "MD"
	ExtJSON    CodeExtension = "JSON"
	ExtYAML    CodeExtension = "YAML"
	ExtGO      CodeExtension = "GO"
	ExtPY      CodeExtension = "PY"
	ExtJAVA    CodeExtension = "JAVA"
	ExtC       CodeExtension = "C"
	ExtCPP     CodeExtension = "CPP"
	ExtRS      CodeExtension = "RS"
	ExtRUBY    CodeExtension = "RB"
	ExtPHP     CodeExtension = "PHP"
	ExtSQL     CodeExtension = "SQL"
	ExtTXT     CodeExtension = "TXT"
	ExtUnknown CodeExtension = "UNKNOWN"
)

type CodeFile struct {
	Type         string        `json:"type"` // "FILE"
	Name         string        `json:"name"`
	FullPath     string        `json:"fullPath"`
	Extension    CodeExtension `json:"extension"`
	IsConflicted bool          `json:"isConflicted"`
}

type CodeDirectory struct {
	Type           string `json:"type"` // "DIRECTORY"
	Name           string `json:"name"`
	FullPath       string `json:"fullPath"`
	SubDirectories []any  `json:"subDirectories,omitempty"` // children: CodeDirectory | CodeFile
}

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

	r.GET("/fileTree", func(c *gin.Context) {
		handleGetFileTree(c, userRepository)
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

// ----- Handler -----

func handleGetFileTree(c *gin.Context, userRepository user.UserRepository) {
	userIDStr := strings.TrimSpace(c.Query("userId"))
	repoName := strings.TrimSpace(c.Query("repoName"))

	if userIDStr == "" || repoName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required query params: userId and repoName"})
		return
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId format"})
		return
	}

	// 1) Get GitHub username from DB
	ghUsername, err := getGithubUsername(c, userRepository, userID)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load user"})
		return
	}
	if ghUsername == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "github username not set for user"})
		return
	}

	// 2) Resolve repo path on disk: repos/{userId}/{githubUsername}/{repoName}
	base := filepath.Join("repos", userID.String())
	repoPath := filepath.Join(base, ghUsername, repoName)

	// Security: clean and ensure inside base
	cleanRepoPath := filepath.Clean(repoPath)
	cleanBase := filepath.Clean(base)
	if !strings.HasPrefix(cleanRepoPath+string(os.PathSeparator), cleanBase+string(os.PathSeparator)) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid path resolution"})
		return
	}

	info, err := os.Stat(cleanRepoPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "repo not found on disk"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read repo path"})
		return
	}
	if !info.IsDir() {
		c.JSON(http.StatusNotFound, gin.H{"error": "repo path is not a directory"})
		return
	}

	// 3) Build **children** of the repo root (array), not a root DIRECTORY node
	children, err := buildRepoChildren(cleanRepoPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build file tree"})
		return
	}

	// Return array of CodeDirectory | CodeFile
	c.JSON(http.StatusOK, children)
}

// ----- Helpers -----

func getGithubUsername(c *gin.Context, userRepository user.UserRepository, userID uuid.UUID) (string, error) {
	u, err := userRepository.GetUserById(c.Request.Context(), userID) // rename if your interface differs
	if err != nil {
		return "", err
	}
	if u == nil {
		return "", os.ErrNotExist
	}
	// Adjust depending on your struct definition (*string vs string)
	return strings.TrimSpace(*u.GithubUsername), nil
}

// buildRepoChildren returns the mixed list of files/dirs directly under repo root
func buildRepoChildren(repoAbsPath string) ([]any, error) {
	entries, err := os.ReadDir(repoAbsPath)
	if err != nil {
		return nil, err
	}

	// Sort: directories first, then files; both alphabetical
	sort.SliceStable(entries, func(i, j int) bool {
		ei, ej := entries[i], entries[j]
		if ei.IsDir() && !ej.IsDir() {
			return true
		}
		if !ei.IsDir() && ej.IsDir() {
			return false
		}
		return strings.ToLower(ei.Name()) < strings.ToLower(ej.Name())
	})

	var children []any

	for _, entry := range entries {
		// Skip .git
		if entry.IsDir() && entry.Name() == ".git" {
			continue
		}

		childAbs := filepath.Join(repoAbsPath, entry.Name())
		childRel := entry.Name() // top-level: rel path is just the entry name

		if entry.IsDir() {
			// Recursively build full directory node
			dirNode, err := buildDirectoryTree(childAbs, filepath.ToSlash(childRel))
			if err != nil {
				return nil, err
			}
			children = append(children, dirNode)
		} else {
			// File node
			fileNode := CodeFile{
				Type:         "FILE",
				Name:         entry.Name(),
				FullPath:     filepath.ToSlash(childRel),
				Extension:    mapExtToCodeExtension(entry.Name()),
				IsConflicted: false,
			}
			children = append(children, fileNode)
		}
	}

	return children, nil
}

// buildDirectoryTree builds a CodeDirectory for a directory (recursively)
func buildDirectoryTree(absPath string, relPath string) (CodeDirectory, error) {
	name := filepath.Base(absPath)
	dirNode := CodeDirectory{
		Type:           "DIRECTORY",
		Name:           name,
		FullPath:       relPath,
		SubDirectories: []any{},
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return dirNode, err
	}

	// Sort: directories first, then files; both alphabetical
	sort.SliceStable(entries, func(i, j int) bool {
		ei, ej := entries[i], entries[j]
		if ei.IsDir() && !ej.IsDir() {
			return true
		}
		if !ei.IsDir() && ej.IsDir() {
			return false
		}
		return strings.ToLower(ei.Name()) < strings.ToLower(ej.Name())
	})

	for _, entry := range entries {
		// Skip .git directory to avoid noise/size
		if entry.IsDir() && entry.Name() == ".git" {
			continue
		}

		childAbs := filepath.Join(absPath, entry.Name())
		childRel := filepath.ToSlash(filepath.Join(relPath, entry.Name()))

		if entry.IsDir() {
			childDir, err := buildDirectoryTree(childAbs, childRel)
			if err != nil {
				return dirNode, err
			}
			dirNode.SubDirectories = append(dirNode.SubDirectories, childDir)
			continue
		}

		fileNode := CodeFile{
			Type:         "FILE",
			Name:         entry.Name(),
			FullPath:     childRel,
			Extension:    mapExtToCodeExtension(entry.Name()),
			IsConflicted: false,
		}
		dirNode.SubDirectories = append(dirNode.SubDirectories, fileNode)
	}

	return dirNode, nil
}

func mapExtToCodeExtension(filename string) CodeExtension {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "js":
		return ExtJS
	case "jsx":
		return ExtJSX
	case "ts":
		return ExtTS
	case "tsx":
		return ExtTSX
	case "html", "htm":
		return ExtHTML
	case "css":
		return ExtCSS
	case "scss":
		return ExtSCSS
	case "md", "mdx":
		return ExtMD
	case "json":
		return ExtJSON
	case "yaml", "yml":
		return ExtYAML
	case "go":
		return ExtGO
	case "py":
		return ExtPY
	case "java":
		return ExtJAVA
	case "c":
		return ExtC
	case "cc", "cpp", "cxx", "hpp", "hh", "hxx":
		return ExtCPP
	case "rs":
		return ExtRS
	case "rb":
		return ExtRUBY
	case "php":
		return ExtPHP
	case "sql":
		return ExtSQL
	case "txt":
		return ExtTXT
	default:
		if ext == "" {
			return ExtUnknown
		}
		return CodeExtension(strings.ToUpper(ext))
	}
}
