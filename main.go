// main.go

package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/siro20/boardstatus/pkg/model"
)

var router *gin.Engine

func main() {

	// for testing only
	u, err := model.GetUserByName("root")
	if u == nil || err != nil {
		var newUser model.User = model.User{
			Username:          "root",
			Name:              "root",
			Password:          "root",
			Email:             "",
			ProfilePictureURL: "",
			OAuthProvider:     "",
		}
		err := newUser.InsertIntoDB()
		if err != nil {
			fmt.Printf("Error inserting user root: %v\n", err)
		}
	} else {
		fmt.Printf("%v\n", u)
	}

	// Set Gin to production mode
	gin.SetMode(gin.ReleaseMode)

	// Set the router as the default one provided by Gin
	router = gin.Default()

	// Process the templates at the start so that they don't have to be loaded
	// from the disk again. This makes serving HTML pages very fast.
	router.LoadHTMLGlob("templates/*")

	// Initialize the routes
	initializeRoutes()

	// Start serving the application
	router.Run()
}

// Render one of HTML, JSON or CSV based on the 'Accept' header of the request
// If the header doesn't specify this, HTML is rendered, provided that
// the template name is present
func render(c *gin.Context, data gin.H, templateName string) {
	loggedInInterface, _ := c.Get("is_logged_in")
	data["is_logged_in"] = loggedInInterface.(bool)

	switch c.Request.Header.Get("Accept") {
	case "application/json":
		// Respond with JSON
		c.JSON(http.StatusOK, data["payload"])
	case "application/xml":
		// Respond with XML
		c.XML(http.StatusOK, data["payload"])
	default:
		// Respond with HTML
		c.HTML(http.StatusOK, templateName, data)
	}
}
