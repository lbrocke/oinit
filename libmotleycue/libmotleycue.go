package libmotleycue

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	ERR_REQUEST              = "http request failed"
	ERR_RESPONSE_BODY        = "cannot parse response body"
	ERR_SERVER_RESPONSE      = "server responded: "
	ERR_SERVER_RESPONSE_CODE = "server responded with unexpected code: %d"
)

type ApiResponseDetail struct {
	Detail string `json:"detail"`
}

type LoginInfo struct {
	Description string `json:"description"`
	LoginHelp   string `json:"login_help"`
	SSHHost     string `json:"ssh_host"`
}

type ApiResponseInfo struct {
	LoginInfo    LoginInfo `json:"login_info"`
	SupportedOPs []string  `json:"supported_OPs"`
}

type Client struct {
	addr string
}

// parseError tries to unmarshal the given response body into
// ApiResponseDetail and returns the enclosed error message as a new error. If
// reading from responseBody or unmarshaling fails, this function return a
// custom error messages.
func parseError(responseBody io.ReadCloser) error {
	var response ApiResponseDetail

	if parseResponse(responseBody, &response) != nil {
		return errors.New(ERR_RESPONSE_BODY)
	}

	return errors.New(response.Detail)
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

// Retrieve service-specific information:
//
//   - login info
//   - supported OPs
func (c Client) GetInfo() (ApiResponseInfo, error) {
	var response ApiResponseInfo

	res, err := http.Get(c.addr + "/info")
	if err != nil {
		return response, errors.New(ERR_REQUEST)
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		return response, parseResponse(res.Body, &response)
	default:
		return response, fmt.Errorf(ERR_SERVER_RESPONSE_CODE, res.StatusCode)
	}
}
