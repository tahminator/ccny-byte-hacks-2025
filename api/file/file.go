package file

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
	"github.com/tahminator/go-react-template/utils"
)

type CodeExtension string

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

	r.GET("/tree/generate", func(c *gin.Context) {
		handleGetFileTree(c, userRepository)
	})

	return r
}

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

	// ------------------------------
	// GIT FLOW: fetch -> merge -> conflict handling (or force conflict mode)
	// ------------------------------

	// git fetch
	if code, _, errOut, err := utils.RunCommand(fmt.Sprintf(`git -C %q fetch`, cleanRepoPath)); err != nil || code != 0 {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "git fetch failed", "details": errOut})
		return
	}

	// Try a straight merge first
	code, _, _, _ := utils.RunCommand(fmt.Sprintf(`git -C %q merge`, cleanRepoPath))
	conflictedMap := map[string]bool{}

	// Always decide based on actual merge state (MERGE_HEAD), not only exit code.
	if isMidMerge(cleanRepoPath) {
		// We are in merge-conflict mode → collect conflicted files
		conflictedMap = collectConflicts(cleanRepoPath)
	} else if code == 0 {
		// Clean merge; nothing conflicted
	} else {
		// Not in merge mode and merge didn't succeed cleanly → force-conflict path

		// 1) Find incoming files
		changed := getChangedFiles(cleanRepoPath) // HEAD..FETCH_HEAD
		paths := splitNonEmptyLines(changed)

		// 2) Stage those paths (if any)
		for _, p := range paths {
			utils.RunCommand(fmt.Sprintf(`git -C %q add -- %q`, cleanRepoPath, p))
		}

		// 3) Marker commit (only if we added something)
		if len(paths) > 0 {
			utils.RunCommand(fmt.Sprintf(`git -C %q commit -m %q`, cleanRepoPath, "This commit is a merge conflict"))
		}

		// 4) Attempt merge to trigger conflicts
		utils.RunCommand(fmt.Sprintf(`git -C %q merge`, cleanRepoPath))

		// 5) If we are now in merge mode, collect conflicts
		if isMidMerge(cleanRepoPath) {
			conflictedMap = collectConflicts(cleanRepoPath)
		}
	}

	// 3) Build **children** of the repo root (array), conflict-aware
	children, err := buildRepoChildren(cleanRepoPath, conflictedMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build file tree"})
		return
	}

	// Return array of CodeDirectory | CodeFile
	c.JSON(http.StatusOK, children)
}

// buildRepoChildren returns the mixed list of files/dirs directly under repo root
func buildRepoChildren(repoAbsPath string, conflicted map[string]bool) ([]any, error) {
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
			dirNode, err := buildDirectoryTree(childAbs, filepath.ToSlash(childRel), conflicted)
			if err != nil {
				return nil, err
			}
			children = append(children, dirNode)
		} else {
			rel := filepath.ToSlash(childRel)
			fileNode := CodeFile{
				Type:         "FILE",
				Name:         entry.Name(),
				FullPath:     rel,
				Extension:    mapExtToCodeExtension(entry.Name()),
				IsConflicted: conflicted[rel],
			}
			children = append(children, fileNode)
		}
	}

	return children, nil
}

// buildDirectoryTree builds a CodeDirectory for a directory (recursively)
func buildDirectoryTree(absPath string, relPath string, conflicted map[string]bool) (CodeDirectory, error) {
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
			childDir, err := buildDirectoryTree(childAbs, childRel, conflicted)
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
			IsConflicted: conflicted[childRel],
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

// getGitDir returns the absolute path to the repo's git dir (handles worktrees)
func getGitDir(repoPath string) (string, error) {
	_, out, _, err := utils.RunCommand(fmt.Sprintf(`git -C %q rev-parse --git-dir`, repoPath))
	if err != nil {
		return "", err
	}
	dir := strings.TrimSpace(out)
	if dir == "" {
		return "", os.ErrNotExist
	}
	// rev-parse may return a relative path like ".git"
	if !filepath.IsAbs(dir) {
		dir = filepath.Join(repoPath, dir)
	}
	return filepath.Clean(dir), nil
}

// isMidMerge returns true iff $GIT_DIR/MERGE_HEAD exists
func isMidMerge(repoPath string) bool {
	gitDir, err := getGitDir(repoPath)
	if err != nil {
		return false
	}
	mergeHead := filepath.Join(gitDir, "MERGE_HEAD")
	if st, err := os.Stat(mergeHead); err == nil && !st.IsDir() {
		return true
	}
	return false
}

// getChangedFiles returns newline-separated files for HEAD..FETCH_HEAD (string)
func getChangedFiles(repoPath string) string {
	_, out, _, _ := utils.RunCommand(fmt.Sprintf(`git -C %q diff --name-only HEAD..FETCH_HEAD`, repoPath))
	return strings.TrimSpace(out)
}

// collectConflicts returns a set of conflicted file paths (relative to repo root)
func collectConflicts(repoPath string) map[string]bool {
	m := map[string]bool{}
	_, out, _, _ := utils.RunCommand(fmt.Sprintf(`git -C %q diff --name-only --diff-filter=U`, repoPath))
	for _, line := range splitNonEmptyLines(out) {
		rel := filepath.ToSlash(strings.TrimSpace(line))
		if rel != "" {
			m[rel] = true
		}
	}
	return m
}

// splitNonEmptyLines trims and splits a blob into non-empty lines
func splitNonEmptyLines(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	lines := strings.Split(s, "\n")
	out := make([]string, 0, len(lines))
	for _, ln := range lines {
		ln = strings.TrimSpace(ln)
		if ln != "" {
			out = append(out, ln)
		}
	}
	return out
}
