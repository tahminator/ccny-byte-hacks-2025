package api

import (

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/tahminator/go-react-template/api/auth"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
	"google.golang.org/genai"
)

func NewRouter(eng *gin.Engine, db *pgxpool.Pool, geminiClient *genai.Client) *gin.RouterGroup {
	r := eng.Group("/api")

	userRepository := user.NewPostgresUserRepository(db)
	sessionRepository := session.NewPostgresSessionRepository(db)

	auth.NewRouter(r, userRepository, sessionRepository)

	return r
}
