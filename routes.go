// routes.go

package main

import (
	// WARN: github.com/gin-contrib/sessions seems broken

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang/glog"
	"github.com/siro20/boardstatus/pkg/model"
	oauth "github.com/siro20/boardstatus/pkg/oauth"
)

func OAuthLoginCallback(c *gin.Context, u oauth.OAuthUser) error {
	var user *model.User
	var err error

	if u.Email == "" {
		user, err = model.GetUserByOAuthLogin(u.Login, u.Provider)
	} else {
		user, err = model.GetUserByEmail(u.Email)
	}

	if user == nil || err != nil {
		var newUser model.User = model.User{
			Username:          u.Login,
			Name:              u.Name,
			Email:             u.Email,
			ProfilePictureURL: u.AvatarURL,
			OAuthProvider:     u.Provider,
		}
		err := newUser.InsertIntoDB()
		if err != nil {
			return err
		}
	}

	session := sessions.Default(c)

	// populate cookie
	session.Set("user", u.Login)
	if err := session.Save(); err != nil {
		glog.Errorf("Failed to save session: %v", err)
	}

	return nil
}

func initializeRoutes() {

	// Use secure cookie store.
	store := sessions.NewCookieStore([]byte("secret"))
	router.Use(sessions.Sessions("mystore", store))

	// Use the setUserStatus middleware for every route to set a flag
	// indicating whether the request was from an authenticated user or not
	router.Use(setUserStatus())

	// Allow BasicAuth for REST API
	//router.Use(BasicAuth())

	// Handle the index route
	router.GET("/", showIndexPage)

	oauth.InstallOAuth2Routers(router, OAuthLoginCallback)

	router.GET("/login", ensureNotLoggedIn(), oauth.ShowOAuth2LoginPage)
	router.GET("/logout", ensureLoggedIn(), oauth.ShowOAuth2LogoutPage)

	// Group user related routes together
	userRoutes2 := router.Group("/u")
	{
		// Handle the GET requests at /u/login
		// Show the login page
		// Ensure that the user is not logged in by using the middleware
		//userRoutes.GET("/login", ensureNotLoggedIn(), showLoginPage)

		// Handle POST requests at /u/login
		// Ensure that the user is not logged in by using the middleware
		userRoutes2.POST("/login", ensureNotLoggedIn(), performLogin)

		// Handle GET requests at /u/logout
		// Ensure that the user is logged in by using the middleware
		//userRoutes.GET("/logout", ensureLoggedIn(), logout)

		// Handle the GET requests at /u/register
		// Show the registration page
		// Ensure that the user is not logged in by using the middleware
		//userRoutes.GET("/register", ensureNotLoggedIn(), showRegistrationPage)

		// Handle POST requests at /u/register
		// Ensure that the user is not logged in by using the middleware
		//userRoutes.POST("/register", ensureNotLoggedIn(), register)
	}

	// Group article related routes together
	boardRoutes := router.Group("/board")
	{
		var b model.Board
		// Handle GET requests at /board/view/id
		boardRoutes.GET("/view/:id", b.RenderShow)

		// Handle the GET requests at /board/create
		// Show the article creation page
		// Ensure that the user is logged in by using the middleware
		boardRoutes.GET("/create", ensureLoggedIn(), showArticleCreationPage)

		// Handle POST requests at /board/create
		// Ensure that the user is logged in by using the middleware
		boardRoutes.POST("/create", ensureLoggedIn(), createArticle)

		// Handle GET requests at /board/list
		boardRoutes.GET("/list/", b.RenderAll)

		boardRoutes.GET("/edit/:id", b.RenderEdit)

	}

	userRoutes := router.Group("/user")
	{
		var u model.User
		// Handle GET requests at /user/view/id
		userRoutes.GET("/view/:id", u.RenderShow)

		// Handle GET requests at /user/list
		userRoutes.GET("/list/", u.RenderAll)

		userRoutes.GET("/edit/:id", u.RenderEdit)

	}

	testsRoutes := router.Group("/test")
	{
		var t model.Test
		// Handle GET requests at /test/view/id
		testsRoutes.GET("/view/:id", t.RenderShow)

		// Handle GET requests at /test/list
		testsRoutes.GET("/list/", t.RenderAll)

		// Handle GET requests at /test/list
		testsRoutes.GET("/edit/:id", t.RenderEdit)
	}
}
