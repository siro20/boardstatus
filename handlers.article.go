// handlers.article.go

package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/siro20/boardstatus/pkg/model"
)

func showIndexPage(c *gin.Context) {
	articles, err := model.GetAllBoards()
	if err != nil {
		return
	}
	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title":   "Board status overview",
		"payload": articles}, "index.html")
}

func showArticleCreationPage(c *gin.Context) {
	// Call the render function with the name of the template to render
	render(c, gin.H{
		"title": "Create New Board"}, "create-board.html")
}

func createArticle(c *gin.Context) {
	// Obtain the POSTed values
	c.Request.ParseForm()
	yaml := "{"

	for key, value := range c.Request.PostForm {
		yaml += key + ": "
		if len(value) == 1 {
			yaml += fmt.Sprintf("%s,", value[0])
		} else {
			yaml += fmt.Sprintf("%s,", value)
		}
	}

	if yaml[len(yaml)-1] == ',' {
		yaml = yaml[0 : len(yaml)-1]
	}
	yaml += "}"

	if b, err := model.CreateNewBoard(yaml); err == nil {
		// If the article is created successfully, show success message
		render(c, gin.H{
			"title":   "Submission Successful",
			"payload": b}, "submission-successful.html")
	} else {
		// if there was an error while creating the article, abort with an error
		c.AbortWithStatus(http.StatusBadRequest)
	}
}
