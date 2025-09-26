package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tahminator/go-react-template/config"
	"github.com/tahminator/go-react-template/database/repository/session"
	"github.com/tahminator/go-react-template/database/repository/user"
	utils "github.com/tahminator/go-react-template/utils"
	"golang.org/x/oauth2"
)

func NewRouter(eng *gin.RouterGroup,
	userRepository user.UserRepository,
	sessionRepository session.SessionRepository,
) *gin.RouterGroup {
	r := eng.Group("/auth")

	r.GET("/validate", func(c *gin.Context) {
		ao, err := utils.ValidateRequest(c, userRepository, sessionRepository)
		if err != nil {
			c.JSON(http.StatusUnauthorized, utils.Failure("You are not authenticated."))
			return
		}
		c.JSON(http.StatusOK, utils.Success("You are authenticated!", ao))
	})

	r.GET("/google", func(c *gin.Context) {
		oauthState := config.GenerateStateOauthCookie()
		c.SetCookie("oauthstate", oauthState, 300, "/", "", os.Getenv("ENV") == "production", true)

		u := config.GetGoogleOAuthConfig().AuthCodeURL(oauthState, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
		c.Redirect(http.StatusTemporaryRedirect, u)
	})

	r.GET("/google/callback", func(c *gin.Context) {
		v, err := c.Cookie("oauthstate")
		if err != nil {
			c.Redirect(http.StatusPermanentRedirect, "/?success=false&message=Failed to authenticate")
			return
		}

		if c.Request.URL.Query().Get("state") != v {
			fmt.Println(fmt.Sprintf("hello %s", c.Request.Form.Get("state")))
			c.Redirect(http.StatusPermanentRedirect, "/?success=false&message=Failed to authenticate")
			return
		}

		data, err := config.GetUserDataFromGoogle(c.Request.URL.Query().Get("code"))
		if err != nil {
			fmt.Println(err)
			c.Redirect(http.StatusPermanentRedirect, "/?success=false&message=Failed to authenticate")
			return
		}

		result := &config.GoogleUserDataObject{}
		err = json.Unmarshal(data, result)
		if err != nil {
			fmt.Println(err)
			c.Redirect(http.StatusPermanentRedirect, "/?success=false&message=Failed to authenticate")
			return
		}

		u, err := userRepository.GetUserByGoogleId(c.Request.Context(), result.Id)
		if u == nil {
			createUser := user.User{
				GoogleId: result.Id,
			}
			u, err = userRepository.CreateUser(c.Request.Context(), &createUser)
			if err != nil {
				fmt.Println(err)
				c.Redirect(http.StatusPermanentRedirect, "/?success=false&message=Failed to authenticate")
				return
			}
		}

		session := &session.Session{
			UserId:    u.Id,
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30),
		}
		session, err = sessionRepository.CreateSession(c.Request.Context(), session)
		if err != nil {
			fmt.Println(err)
			c.Redirect(http.StatusPermanentRedirect, "/?success=false&message=Failed to authenticate")
			return
		}

		ttl := time.Until(session.ExpiresAt)

		c.SetCookie("session", session.Id.String(), int(ttl.Seconds()), "/", "", os.Getenv("ENV") == "production", true)

		c.Redirect(http.StatusPermanentRedirect, "/?success=true&message=You have been authenticated!")
	})
	return r
}
