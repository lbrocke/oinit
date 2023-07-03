package api

import (
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/lbrocke/oinit/internal/caconfig"
	"github.com/lbrocke/oinit/pkg/libmotleycue"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/ssh"
	"golang.org/x/exp/slices"
)

const (
	API_VERSION = "0.1.0"

	ERR_BAD_BODY       = "Request body is malformed."
	ERR_UNKNOWN_HOST   = "Unknown host."
	ERR_GATEWAY_DOWN   = "motley_cue is not reachable."
	ERR_UNAUTHORIZED   = "User is not authorized or suspended."
	ERR_INTERNAL_ERROR = "Internal server error."
)

type ApiResponseError struct {
	Error string `json:"error"`
}

type ApiResponseIndex struct {
	Version string `json:"version"`
}

type ApiResponseHost struct {
	PublicKey string     `json:"publickey"`
	Providers []Provider `json:"providers"`
}

type ApiResponseCertificate struct {
	Certificate string `json:"certificate"`
}

type Provider struct {
	URL    string   `json:"url"`
	Scopes []string `json:"scopes"`
}

type UriHost struct {
	Host string `uri:"host" binding:"required"`
}

type FormHostCertificate struct {
	Pubkey string `json:"pubkey" binding:"required"`
	Token  string `json:"token" binding:"required"`
}

func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, ApiResponseError{
		Error: msg,
	})
}

type customLog struct {
}

// Custom log format that imitates gin's output
func (writer customLog) Write(bytes []byte) (int, error) {
	return fmt.Print("[API] " + time.Now().Format("2006/01/02 - 15:04:05") + " " + string(bytes))
}

type providerCache struct {
	hosts map[string]providerCacheEntry
}

type providerCacheEntry struct {
	providers []Provider
	expires   time.Time
}

func (c providerCache) Get(host string) ([]Provider, bool) {
	var providers []Provider

	entry, ok := c.hosts[host]
	if !ok {
		return providers, false
	}

	if time.Now().After(entry.expires) {
		delete(c.hosts, host)

		return providers, false
	}

	return entry.providers, true
}

func (c providerCache) Set(host string, providers []Provider) {
	c.hosts[host] = providerCacheEntry{
		providers: providers,
		expires:   time.Now().Add(10 * time.Minute),
	}
}

var cache = providerCache{
	hosts: make(map[string]providerCacheEntry),
}

// GetIndex is the handler for GET /
//
//	@Summary		Get API version
//	@Description	Return the running API version.
//	@Produce		json
//	@Success		200	{object}	ApiResponseIndex
//	@Router			/ [get]
func GetIndex(c *gin.Context) {
	c.JSON(http.StatusOK, ApiResponseIndex{
		Version: API_VERSION,
	})
}

// GetHost is the handler for GET /:host
//
//	@Summary		Get host information
//	@Description	Return the CA public key and supported OpenID Connect providers with their required scopes.
//	@Produce		json
//	@Param			host	path		string	true	"Host"	example("example.com")
//	@Success		200		{object}	ApiResponseHost
//	@Failure		400		{object}	ApiResponseError
//	@Failure		500		{object}	ApiResponseError
//	@Failure		502		{object}	ApiResponseError
//	@Router			/{host} [get]
func GetHost(c *gin.Context) {
	var host UriHost

	if c.ShouldBindUri(&host) != nil {
		Error(c, http.StatusBadRequest, ERR_BAD_BODY)
		return
	}

	host.Host = strings.ToLower(host.Host)

	conf, ok := c.MustGet("config").(caconfig.Config)
	if !ok {
		Error(c, http.StatusInternalServerError, ERR_INTERNAL_ERROR)
		return
	}

	info, err := conf.GetInfo(host.Host)
	if err != nil {
		Error(c, http.StatusBadRequest, ERR_UNKNOWN_HOST)
		return
	}

	providers, ok := cache.Get(info.URL)
	if !ok {
		hostInfo, err := libmotleycue.NewClient(info.URL).GetInfo()
		if err != nil {
			Error(c, http.StatusBadGateway, ERR_GATEWAY_DOWN)
			return
		}

		// Iterate OpsInfo instead of SupportedOPs to only add hosts for which
		// scopes are defined. Validate that issuer is listed in SupportedOPs
		// however.
		for issuer, info := range hostInfo.OpsInfo {
			if slices.Contains(hostInfo.SupportedOPs, issuer) {
				providers = append(providers, Provider{
					URL:    issuer,
					Scopes: info.Scopes,
				})
			}
		}

		cache.Set(info.URL, providers)
	}

	c.JSON(http.StatusOK, ApiResponseHost{
		PublicKey: strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(info.HostCAPublicKey)), "\n"),
		Providers: providers,
	})
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
//	@Failure		401		{object}	ApiResponseError
//	@Failure		500		{object}	ApiResponseError
//	@Failure		502		{object}	ApiResponseError
//	@Router			/{host}/certificate [post]
func PostHostCertificate(c *gin.Context) {
	log.SetFlags(0)
	log.SetOutput(new(customLog))

	var host UriHost
	var body FormHostCertificate

	if c.ShouldBindUri(&host) != nil || c.ShouldBindJSON(&body) != nil {
		Error(c, http.StatusBadRequest, ERR_BAD_BODY)
		return
	}

	host.Host = strings.ToLower(host.Host)

	pubkey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(body.Pubkey))
	if err != nil {
		Error(c, http.StatusBadRequest, ERR_BAD_BODY)
		return
	}

	conf, ok := c.MustGet("config").(caconfig.Config)
	if !ok {
		Error(c, http.StatusInternalServerError, ERR_INTERNAL_ERROR)
		return
	}

	info, err := conf.GetInfo(host.Host)
	if err != nil {
		Error(c, http.StatusBadRequest, ERR_UNKNOWN_HOST)
		return
	}

	// Parse JWT without verifying it, as the signer key is unknown to the CA.
	// motley_cue will verify the token instead.
	token, _, err := new(jwt.Parser).ParseUnverified(body.Token, jwt.MapClaims{})
	if err != nil {
		Error(c, http.StatusBadRequest, ERR_BAD_BODY)
		return
	}

	status, err := libmotleycue.NewClient(info.URL).GetUserDeploy(body.Token)
	if err != nil || status.State != libmotleycue.StateDeployed {
		// Either something went wrong with the HTTP request/deployment, the
		// access token is not valid (e.g. expired) or the user is suspended.
		Error(c, http.StatusUnauthorized, ERR_UNAUTHORIZED)
		return
	}

	certDuration := info.CertDuration
	// If CertDuration is set to 0 or negative number, use the expiry date of the
	// given token as "valid before" date.
	if certDuration <= 0 {
		if exp, err := token.Claims.GetExpirationTime(); err == nil {
			certDuration = uint64(time.Until(exp.Time).Seconds())
		}
	}

	cert := generateUserCertificate(host.Host, pubkey, status.Credentials.SSHUser, certDuration)

	signer, err := ssh.NewSignerFromKey(info.UserCAPrivateKey)
	if err != nil || cert.SignCert(rand.Reader, signer) != nil {
		Error(c, http.StatusUnauthorized, ERR_INTERNAL_ERROR)
		return
	}

	// Make sure that certificate is valid and (this *very* is important!) has
	// the force-command option set to the correct (non-empty) value.
	if !validateUserCertificate(cert, info.UserCAPublicKey) {
		Error(c, http.StatusInternalServerError, ERR_INTERNAL_ERROR)
		return
	}

	log.Printf("Issued certificate '%s' valid until '%s'", ssh.FingerprintSHA256(cert.Key), time.Unix(int64(cert.ValidBefore-1), 0))

	c.JSON(http.StatusCreated, ApiResponseCertificate{
		Certificate: strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(&cert)), "\n"),
	})
}
