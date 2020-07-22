package oauth

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siro20/boardstatus/pkg/helper"
)

type OAuth2 interface {
	Name() string
	LoginHandlerURL(token string) string
	AuthHandlerURL() string
	AuthHandler(c *gin.Context)
	LoginHandler(c *gin.Context)
}

type OAuthUser struct {
	Login     string
	Name      string
	Email     string
	AvatarURL string
	Provider  string
}

type OAuthCallback func(*gin.Context, OAuthUser) error

func OAuthRandToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

func OAuthSessionToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	hasher := md5.New()
	hasher.Write(b)
	return hex.EncodeToString(hasher.Sum(nil))
}

var providers []OAuth2

func OAuthGetRandToken(c *gin.Context) string {
	// Handle the exchange code to initiate a transport.
	session := sessions.Default(c)

	ret, ok := session.Get("state").(string)
	if !ok {
		fmt.Printf("Internal error, state is not a string\n")
		return ""
	}
	fmt.Printf("OAuthGetRandToken %s\n", ret)

	return ret
}

// This middleware sets the session token used by OAuth
func setOAuthRandToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		retrievedState, ok := session.Get("state").(string)
		if !ok || retrievedState == "" {
			token := OAuthRandToken()
			session := sessions.Default(c)
			session.Set("state", token)
			session.Options(sessions.Options{
				MaxAge:   300,
				HttpOnly: true,
			})
			session.Save()
		}
	}
}

func InstallOAuth2Routers(router *gin.Engine, f OAuthCallback) {

	router.Use(setOAuthRandToken())

	p, err := InitOAuth2Github(f)
	if err == nil {
		providers = append(providers, p)
		router.GET(p.AuthHandlerURL(),
			p.AuthHandler)
	} else {
		fmt.Printf("Error starting OAuth2 provider %v\n", err)
	}

	p, err = InitOAuth2Google(f)
	if err == nil {
		providers = append(providers, p)
		router.GET(p.AuthHandlerURL(),
			p.AuthHandler)
	} else {
		fmt.Printf("Error starting OAuth2 provider %v\n", err)
	}

}

func ShowOAuth2LoginPage(c *gin.Context) {
	var lp []map[string]string

	token := OAuthGetRandToken(c)
	// Add the supported providers to the login page
	for i := range providers {
		lp = append(lp, map[string]string{
			"OAuthURL": providers[i].LoginHandlerURL(token),
			"Name":     providers[i].Name()})

	}

	// Call the render function with the name of the template to render
	helper.Render(c, gin.H{
		"title":          "Login",
		"LoginProviders": lp,
	}, "login.html")
}

func ShowOAuth2LogoutPage(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()

	c.Set("oauth_session_token", "")

	// Redirect to the home page
	c.Redirect(http.StatusTemporaryRedirect, "/")
}
