package main

import (
	"log"
	docs "oinit/api/docs"
	"oinit/internal/api"
	"oinit/internal/caconfig"
	"os"

	"github.com/gin-gonic/gin"
)

const (
	USAGE = "Usage: oinit-ca <host:port> <path/to/config>"

	SWAGGER_TITLE = "oinit CA API"
	SWAGGER_DESC  = "Swagger documentation for the oinit CA REST API."
)

func ConfigMiddleware(config caconfig.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("config", config)
		c.Next()
	}
}

func main() {
	args := os.Args[1:]
	if len(args) != 2 {
		log.Fatalln(USAGE)
	}

	addr := args[0]
	conf := args[1]

	cfg, err := caconfig.LoadConfig(conf)
	if err != nil {
		log.Fatalln("Error while loading config: " + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.Default()
	router.Use(ConfigMiddleware(cfg))

	gAPI := router.Group("/api")
	{
		gAPI.GET("/docs/*any", api.GetSwagger)

		v1 := gAPI.Group("/v1")
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
	}

	docs.SwaggerInfo.Version = api.API_VERSION
	docs.SwaggerInfo.BasePath = "/api/v1"
	docs.SwaggerInfo.Title = SWAGGER_TITLE
	docs.SwaggerInfo.Description = SWAGGER_DESC

	router.Run(addr)
}
