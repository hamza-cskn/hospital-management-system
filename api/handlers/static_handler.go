package handlers

import (
	"github.com/gin-gonic/gin"
)

// GetStaticPageHandler returns a handler function for serving static pages
func GetStaticPageHandler(page string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.File("./frontend/" + page)
	}
}


