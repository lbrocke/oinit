package api

import (
	"github.com/gin-gonic/gin"
)

const (
	API_VERSION = "0.1.0"
)

type ApiResponseError struct {
	Error string `json:"error"`
}

type ApiResponseIndex struct {
	Version string `json:"version"`
}

type ApiResponseHost struct {
	PublicKey string   `json:"publickey"`
	Providers []string `json:"providers"`
}

type ApiResponseCertificate struct {
	Certificate string `json:"certificate"`
}

type UriHost struct {
	Host string `uri:"host" binding:"required"`
}

type FormHostCertificate struct {
	Pubkey string `json:"pubkey" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

// GetIndex is the handler for GET /
//
//	@Summary		Get API version
//	@Description	Return the running API version.
//	@Produce		json
//	@Success		200	{object}	ApiResponseIndex
//	@Router			/ [get]
func GetIndex(c *gin.Context) {
	// todo
}

// GetHost is the handler for GET /:host
//
//	@Summary		Get host information
//	@Description	Return the CA public key and supported OpenID Connect providers.
//	@Produce		json
//	@Param			host	path		string	true	"Host"	example("example.com")
//	@Success		200		{object}	ApiResponseHost
//	@Failure		400		{object}	ApiResponseError
//	@Router			/{host} [get]
func GetHost(c *gin.Context) {
	// todo
}

// PostHostCertificate is the handler for POST /:host/certificate
//
//	@Summary		Generate SSH certificate
//	@Description	Generate and return a new SSH certificate using the given public key and access token.
//	@Accept			json
//	@Produce		json
//	@Param			host	path		string				true	"Host"	example("example.com")
//	@Param			body	body		FormHostCertificate	true	"Public key and access token"
//	@Success		201		{object}	ApiResponseCertificate
//	@Failure		400		{object}	ApiResponseError
//	@Router			/{host}/certificate [post]
func PostHostCertificate(c *gin.Context) {
	// todo
}
