package oauth

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"net/url"

	"github.com/siro20/boardstatus/pkg/helper"
)

// Credentials which stores google ids.
type OAuth2GoogleCredentials struct {
	Web struct {
		Cid       string   `json:"client_id"`
		Csecret   string   `json:"client_secret"`
		RedisURIs []string `json:"redirect_uris"`
	} `json:"web"`
}

// User is a retrieved and authentiacted user.
type OAuth2GoogleUser struct {
	Sub           string `json:"sub"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Profile       string `json:"profile"`
	Picture       string `json:"picture"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Locale        string `json:"locale"`
	Gender        string `json:"gender"`
}

type OAuth2Google struct {
	cred     OAuth2GoogleCredentials
	conf     oauth2.Config
	user     *OAuth2GoogleUser
	callback OAuthCallback
}

func InitOAuth2Google(f OAuthCallback) (OAuth2, error) {
	var o OAuth2Google
	file, err := ioutil.ReadFile("./google.creds.json")
	if err != nil {
		return nil, fmt.Errorf("File error: %v\n", err)
	}
	err = json.Unmarshal(file, &o.cred)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse API credentials %v\n", err)
	}

	o.conf.ClientID = o.cred.Web.Cid
	o.conf.ClientSecret = o.cred.Web.Csecret
	o.conf.RedirectURL = o.cred.Web.RedisURIs[0]
	o.conf.Endpoint = google.Endpoint
	o.conf.Scopes = []string{
		"https://www.googleapis.com/auth/userinfo.email",
		"https://www.googleapis.com/auth/userinfo.profile",
	}

	o.callback = f
	fmt.Printf("Oauth login handler url %s\n", o.cred.Web.RedisURIs[0])
	return o, nil
}

func (o OAuth2Google) LoginHandler(c *gin.Context) {
	token := OAuthGetRandToken(c)
	c.Redirect(http.StatusFound, o.LoginHandlerURL(token))
	c.Abort()
}

func (o OAuth2Google) Name() string {
	return "Google"
}

func (o OAuth2Google) AuthHandler(c *gin.Context) {
	// Handle the exchange code to initiate a transport.
	token := OAuthGetRandToken(c)
	fmt.Printf("Googl auth handler\n")
	if token != c.Query("state") {
		c.AbortWithError(http.StatusUnauthorized, fmt.Errorf("Invalid session state: %v %v", token, c.Query("state")))
		return
	}

	tok, err := o.conf.Exchange(oauth2.NoContext, c.Query("code"))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	client := o.conf.Client(oauth2.NoContext, tok)
	email, err := client.Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	defer email.Body.Close()
	data, _ := ioutil.ReadAll(email.Body)

	var u OAuth2GoogleUser
	err = json.Unmarshal([]byte(data), &u)
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var user OAuthUser
	user.Login = u.Sub
	user.Provider = "google"
	user.Name = u.Name
	user.Email = u.Email
	user.AvatarURL = u.Picture

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

func (o OAuth2Google) LoginHandlerURL(token string) string {

	return o.conf.AuthCodeURL(token)
}
func (o OAuth2Google) AuthHandlerURL() string {
	if len(o.cred.Web.RedisURIs) > 0 {
		u, err := url.Parse(o.cred.Web.RedisURIs[0])
		if err != nil {
			return ""
		}
		return u.Path
	} else {
		return ""
	}
}
