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
	ERR_UNEXPECTED_ERROR     = "server returned error code but no description"
	ERR_SERVER_RESPONSE      = "server responded: "
	ERR_SERVER_RESPONSE_CODE = "server responded with code: %d"
)

type ApiResponseDetail struct {
	Detail string `json:"detail"`
}

type LoginInfo struct {
	Description string `json:"description"`
	LoginHelp   string `json:"login_help"`
	SSHHost     string `json:"ssh_host"`
}

type Credentials struct {
	CommandLine string `json:"commandline"`
	Description string `json:"description"`
	LoginHelp   string `json:"login_help"`
	SSHHost     string `json:"ssh_host"`
	SSHUser     string `json:"ssh_user"`
}

type ApiResponseInfo struct {
	LoginInfo    LoginInfo `json:"login_info"`
	SupportedOPs []string  `json:"supported_OPs"`
}

type UserStatusState string

const (
	StateDeployed    UserStatusState = "deployed"
	StateNotDeployed UserStatusState = "not_deployed"
	StateSuspended   UserStatusState = "suspended"
)

// Also called "FeudalResponse" in API docs
type ApiResponseUserStatus struct {
	State       UserStatusState `json:"state"`
	Message     string          `json:"message"`
	Credentials Credentials     `json:"credentials"`
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

	// Make sure the .Detail field was filled after Unmarshalling the JSON data.
	if response.Detail == "" {
		response.Detail = ERR_UNEXPECTED_ERROR
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

// GetInfo calls GET /info.
//
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

// getUser is the implementation of both GET /user/get_status and GET
// /user/deploy, as their request parameters and response are identical.
func (c Client) getUser(path string, token string) (ApiResponseUserStatus, error) {
	var response ApiResponseUserStatus

	req, err := http.NewRequest(http.MethodGet, c.addr+path, nil)
	if err != nil {
		return response, errors.New(ERR_REQUEST)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return response, errors.New(ERR_REQUEST)
	}

	defer res.Body.Close()

	switch res.StatusCode {
	case 200:
		return response, parseResponse(res.Body, &response)
	case 401:
		fallthrough
	case 403:
		fallthrough
	case 404:
		return response, parseError(res.Body)
	case 422:
		// In this case, the response body has a different structure and cannot
		// be parsed easily into a ApiResponseDetail struct, therefore return
		// custom error.
		fallthrough
	default:
		return response, fmt.Errorf(ERR_SERVER_RESPONSE_CODE, res.StatusCode)
	}
}

// GetUserStatus calls GET /user/status.
//
// Get information about your local account:
//
//   - state: one of the supported states, such as deployed, not_deployed, suspended.
//   - message: could contain additional information, such as the local username
//
// Requires an authorised user.
func (c Client) GetUserStatus(token string) (ApiResponseUserStatus, error) {
	return c.getUser("/user/get_status", token)
}

// GetUserDeploy calls GET /user/deploy.
//
// Provision a local account.
// Requires an authorised user.
func (c Client) GetUserDeploy(token string) (ApiResponseUserStatus, error) {
	return c.getUser("/user/deploy", token)
}
