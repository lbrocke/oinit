package oidc

import (
	"net"
	"os"

	"github.com/indigo-dc/liboidcagent-go"
)

type socket struct {
	AddressEnvVar string
	Type          string
}

func AgentIsRunning() bool {
	// From https://github.com/indigo-dc/liboidcagent-go/blob/master/comm.go
	sockets := []socket{
		{
			"OIDC_SOCK",
			"unix",
		},
		{
			"OIDC_REMOTE_SOCK",
			"tcp",
		},
	}

	for _, sock := range sockets {
		if _, err := net.Dial(sock.Type, os.Getenv(sock.AddressEnvVar)); err == nil {
			return true
		}
	}

	return false
}

func GetToken(issuer string) (string, error) {
	req := liboidcagent.TokenRequest{
		IssuerURL:       issuer,
		Scopes:          []string{"openid", "profile", "offline_access"},
		ApplicationHint: "oinit",
	}

	res, err := liboidcagent.GetTokenResponse(req)
	if err != nil {
		return "", err
	}

	return res.Token, err
}
