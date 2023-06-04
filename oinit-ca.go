package main

import (
	"log"
	"oinit-ca/api"
	"oinit-ca/docs"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	USAGE = "Usage: oinit-ca <host>[:port]"

	SWAGGER_TITLE = "oinit CA API"
	SWAGGER_DESC  = "Swagger documentation for the oinit CA REST API."
)

func main() {
	args := os.Args[1:]
	if len(args) != 1 {
		log.Fatalln(USAGE)
	}

	addr := args[0]

	router := gin.Default()

	v1 := router.Group("/api/v1")
	{
		v1.GET("/", api.GetIndex)

		v1.GET("/:host", api.GetHost)
		// Although from the client perspective this route _gets_ a certificate, it
		//  a) generates a new certificate every time (and thus is not cacheable), and
		//  b) must accept an access token (which is a sensitive information better
		//     transmitted in the request body, not as query parameter).
		// Therefore this route uses the POST method rather then GET.
		v1.POST("/:host/certificate", api.PostHostCertificate)
	}

	docs.SwaggerInfo.Version = api.API_VERSION
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Host = addr
	docs.SwaggerInfo.Title = SWAGGER_TITLE
	docs.SwaggerInfo.Description = SWAGGER_DESC

	router.GET("/docs/*any", api.GetSwagger)

	gin.SetMode(gin.ReleaseMode)
	router.Run(addr)
}
