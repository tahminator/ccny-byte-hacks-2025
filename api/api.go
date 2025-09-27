package api

import (
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tahminator/go-react-template/api/auth"
	"github.com/tahminator/go-react-template/api/gemini"
	"github.com/tahminator/go-react-template/api/github"
	"github.com/tahminator/go-react-template/database/repository/repo_chunks"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
	"google.golang.org/genai"
)

func NewRouter(eng *gin.Engine, db *pgxpool.Pool, geminiClient *genai.Client) *gin.RouterGroup {
	r := eng.Group("/api")

	userRepository := user.NewPostgresUserRepository(db)
	sessionRepository := session.NewPostgresSessionRepository(db)
	repoChunksRepository := repo_chunks.NewPostgresRepoChunksRepository(db)

	auth.NewRouter(r, userRepository, sessionRepository)
	gemini.NewRouter(r, geminiClient, repoChunksRepository)
	github.NewRouter(r, userRepository, sessionRepository)

	return r
}
