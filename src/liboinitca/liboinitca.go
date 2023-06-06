package liboinitca

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	ERR_REQUEST              = "http request failed"
	ERR_RESPONSE_BODY        = "cannot parse response body"
	ERR_SERVER_RESPONSE      = "server responded: "
	ERR_SERVER_RESPONSE_CODE = "server responded with unexpected code: %d"
)

type Client struct {
	addr string
}

type ApiResponseError struct {
	Error string `json:"error"`
}

type ApiResponseHost struct {
	PublicKey string   `json:"publickey"`
	Providers []string `json:"providers"`
}

type ApiResponseCertificate struct {
	Certificate string `json:"certificate"`
}

type FormHostCertificate struct {
	Pubkey string `json:"pubkey"`
	Token  string `json:"token"`
}

// parseError tries to unmarshal the given response body into
// ApiResponseError and returns the enclosed error message as a new error. If
// reading from responseBody or unmarshaling fails, this function return a
// custom error messages.
func parseError(responseBody io.ReadCloser) error {
	var response ApiResponseError

	if parseResponse(responseBody, &response) != nil {
		return errors.New(ERR_RESPONSE_BODY)
	}

	return errors.New(response.Error)
}

// parseResponse tries to unmarshal the given response body into a given struct.
// An error is returned for reader or unmarshalling errors.
func parseResponse(responseBody io.ReadCloser, into interface{}) error {
	if body, err := io.ReadAll(responseBody); err != nil ||
		json.Unmarshal(body, into) != nil {
		return errors.New(ERR_RESPONSE_BODY)
	}

	return nil
}

// NewClient creates a new API client. addr is the server address (and port)
// including the protocol, such as http://example.com:8080
func NewClient(addr string) Client {
	addr, _ = strings.CutSuffix(addr, "/")

	return Client{
		addr: addr,
	}
}

// Return the CA public key and supported OpenID Connect providers.
func (c Client) GetHost(host string) (ApiResponseHost, error) {
	var response ApiResponseHost

	res, err := http.Get(c.addr + "/" + url.PathEscape(host))
	if err != nil {
		return response, errors.New(ERR_REQUEST)
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		return response, parseResponse(res.Body, &response)
	case 400:
		return response, parseError(res.Body)
	default:
		return response, fmt.Errorf(ERR_SERVER_RESPONSE_CODE, res.StatusCode)
	}
}

// Generate and return a new SSH certificate using the given access token.
func (c Client) PostHostCertificate(host, pubkey, token string) (ApiResponseCertificate, error) {
	var response ApiResponseCertificate

	reqBody, err := json.Marshal(FormHostCertificate{
		Pubkey: pubkey,
		Token:  token,
	})
	if err != nil {
		return response, err
	}

	res, err := http.Post(c.addr+"/"+host+"/"+url.PathEscape("certificate"), "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return response, errors.New(ERR_REQUEST)
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 201:
		return response, parseResponse(res.Body, &response)
	case 400:
		return response, parseError(res.Body)
	default:
		return response, fmt.Errorf(ERR_SERVER_RESPONSE_CODE, res.StatusCode)
	}
}
