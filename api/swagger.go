package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func GetSwagger(c *gin.Context) {
	if c.Param("any") == "/" {
		c.Redirect(http.StatusMovedPermanently, "/docs/index.html")
		return
	}

	ginSwagger.WrapHandler(swaggerFiles.Handler)(c)
}
