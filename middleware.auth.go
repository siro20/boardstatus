// middleware.auth.go

package main

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/siro20/boardstatus/pkg/model"
)

func authorizationHeader(user, password string) string {
	base := user + ":" + password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(base))
}

func BasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		var header string

		if values, _ := c.Request.Header["Authorization"]; len(values) > 0 {
			header = values[0]
		}
		if header == "" {
			c.AbortWithStatus(http.StatusBadRequest)
			return
		}

		// Search user in the slice of allowed credentials
		u, err := model.GetUserByBasicAuth(header)
		if u == nil || err != nil {
			// Credentials doesn't match, we return 401 and abort handlers chain.
			c.Header("WWW-Authenticate", "Basic realm="+strconv.Quote("Authorization Required"))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("user", u)
	}
}

// This middleware ensures that a request will be aborted with an error
// if the user is not logged in
func ensureLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If there's an error or if the token is empty
		// the user is not logged in
		loggedInInterface, exists := c.Get("is_logged_in")

		if exists {
			loggedIn, ok := loggedInInterface.(bool)
			if ok && !loggedIn {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		}
	}
}

// This middleware ensures that a request will be aborted with an error
// if the user is already logged in
func ensureNotLoggedIn() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If there's no error or if the token is not empty
		// the user is already logged in
		loggedInInterface, exists := c.Get("is_logged_in")
		if exists {
			loggedIn, ok := loggedInInterface.(bool)
			if ok && loggedIn {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
		}
	}
}

// This middleware sets whether the user is logged in or not
func setUserStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		session := sessions.Default(c)
		user := session.Get("user")
		if user != nil {
			u, ok := user.(string)
			c.Set("is_logged_in", ok && u != "")
			fmt.Printf("is_logged_in %v %s\n", ok && u != "", u)

		} else {
			fmt.Printf("is_logged_in false\n")
			c.Set("is_logged_in", false)
		}
	}
}
