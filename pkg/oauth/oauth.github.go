package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"net/url"

	"github.com/gin-gonic/gin"
	gogithub "github.com/google/go-github/github"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"

	"github.com/siro20/boardstatus/pkg/helper"
)

// Credentials which stores github ids.
type OAuth2GithubCredentials struct {
	Cid       string   `json:"client_id"`
	Csecret   string   `json:"client_secret"`
	RedisURIs []string `json:"redirect_uris"`
}

type OAuth2Github struct {
	cred     OAuth2GithubCredentials
	conf     oauth2.Config
	callback OAuthCallback
}

func InitOAuth2Github(f OAuthCallback) (OAuth2, error) {
	var o OAuth2Github
	file, err := ioutil.ReadFile("./github.creds.json")
	if err != nil {
		return nil, fmt.Errorf("File error: %v\n", err)
	}
	err = json.Unmarshal(file, &o.cred)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse API credentials %v\n", err)
	}

	o.conf.ClientID = o.cred.Cid
	o.conf.ClientSecret = o.cred.Csecret
	o.conf.RedirectURL = o.cred.RedisURIs[0]
	o.conf.Endpoint = github.Endpoint
	o.conf.Scopes = []string{"user:email", "read:user"}

	o.callback = f
	fmt.Printf("Github Oauth login handler url %s\n", o.cred.RedisURIs[0])

	return o, nil
}

func (o OAuth2Github) LoginHandler(c *gin.Context) {
	fmt.Printf("LoginHandler\n")

	token := OAuthGetRandToken(c)
	c.Redirect(http.StatusFound, o.LoginHandlerURL(token))
	c.Abort()
}

func (o OAuth2Github) Name() string {
	return "Github"
}

func (o OAuth2Github) AuthHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	token := OAuthGetRandToken(c)

	if c.Query("error") != "" {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": c.Query("error")})
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Auth querry error: %s", c.Query("error_description")))
		return
	}

	if token != c.Query("state") {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Invalid session state"})
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %v %v", token, c.Query("state")))
		return
	}

	tok, err := o.conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": err.Error()})
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	oauthClient := o.conf.Client(oauth2.NoContext, tok)

	client := gogithub.NewClient(oauthClient)
	// fetch myself
	u, _, err := client.Users.Get(oauth2.NoContext, "")
	if err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": err.Error()})
		fmt.Printf("client.Users.Get() faled with '%s'\n", err)
		return
	}
	if u.Login == nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": "Invalid data"})
		return
	}
	fmt.Printf("Logged in as GitHub user: %s\n", *u.Login)

	var user OAuthUser
	user.Login = *u.Login
	user.Provider = "github"
	if u.Name != nil && *u.Name != "" {
		user.Name = *u.Name
	}
	if u.Email != nil && *u.Email != "" {
		user.Email = *u.Email
	}
	if u.AvatarURL != nil && *u.AvatarURL != "" {
		user.AvatarURL = *u.AvatarURL
	}

	err = o.callback(c, user)

	if err != nil {
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"ErrorTitle":   "Login Failed",
			"ErrorMessage": err.Error()})
	} else {
		helper.Render(c, gin.H{
			"title": "Successful Login"}, "login-successful.html")

		c.Status(http.StatusOK)
	}
}

func (o OAuth2Github) LoginHandlerURL(token string) string {

	return o.conf.AuthCodeURL(token)
}

func (o OAuth2Github) AuthHandlerURL() string {
	if len(o.cred.RedisURIs) > 0 {
		u, err := url.Parse(o.cred.RedisURIs[0])
		if err != nil {
			return ""
		}
		return u.Path
	} else {
		return ""
	}
}
